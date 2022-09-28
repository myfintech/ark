package build

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/moby/buildkit/session/secrets/secretsprovider"
	"github.com/moby/buildkit/util/appcontext"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stringid"
	"github.com/moby/buildkit/session/filesync"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"golang.org/x/sync/errgroup"

	"github.com/hashicorp/hcl/v2"
	vault "github.com/hashicorp/vault/api"
	"github.com/zclconf/go-cty/cty"

	"github.com/myfintech/ark/src/go/lib/log"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/fs"
	"github.com/myfintech/ark/src/go/lib/hclutils"
	"github.com/myfintech/ark/src/go/lib/utils"
)

const (
	ARK_LLB_FILE     = "__ark.llb"
	ARK_CONTEXT_FILE = "context.tar.gz"

	arkEPCopyTemplate = `COPY --from=ark-entrypoint /ark-entrypoint-linux /usr/local/bin/ark-entrypoint
ENV ARK_TARGET_ADDRESS=%s
ENV ARK_TARGET_HASH=%s
`
)

var (
	BUILDKIT_INLINE_CACHE = "1"
	appCTX                = appcontext.Context()
)

// Target defines the required and optional attributes for defining a Docker Image Build
type Target struct {
	*base.RawTarget            `json:"-"`
	Repo                       hcl.Expression `hcl:"repo,attr"`
	Dockerfile                 hcl.Expression `hcl:"dockerfile,attr"` // Has to be relative to the repo/workspace root
	BuildArgs                  hcl.Expression `hcl:"build_args,optional"`
	Target                     hcl.Expression `hcl:"target,optional"`
	Tags                       hcl.Expression `hcl:"tags,attr"`
	Output                     hcl.Expression `hcl:"output,attr"`
	DumpContext                hcl.Expression `hcl:"dump_context,attr"`
	CacheInline                hcl.Expression `hcl:"cache_inline,optional"`
	CacheFrom                  hcl.Expression `hcl:"cache_from,optional"`
	Secrets                    hcl.Expression `hcl:"secrets,optional"`
	DisableEntrypointInjection hcl.Expression `hcl:"disable_entrypoint_injection,optional"`
}

// ComputedAttrs used to store the computed attributes of a docker_image target
type ComputedAttrs struct {
	Repo                       string             `hcl:"repo,attr"`
	Dockerfile                 string             `hcl:"dockerfile,attr"` // Has to be relative to the repo/workspace root
	BuildArgs                  map[string]*string `hcl:"build_args,optional"`
	Target                     string             `hcl:"target,optional"`
	Output                     string             `hcl:"output,optional"`
	DumpContext                bool               `hcl:"dump_context,optional"`
	Tags                       []string           `hcl:"tags,optional"`
	CacheFrom                  []string           `hcl:"cache_from,optional"`
	CacheInline                bool               `hcl:"cache_inline,optional"`
	Secrets                    map[string]string  `hcl:"secrets,optional"` // formatted as secretID: vaultPath
	DisableEntrypointInjection bool               `hcl:"disable_entrypoint_injection,optional"`
}

// Attributes return combined rawTarget.Attributes with typedTarget.Attributes.
func (t Target) Attributes() map[string]cty.Value {
	computed := t.ComputedAttrs()
	return hclutils.MergeMapStringCtyValue(t.RawTarget.Attributes(), map[string]cty.Value{
		"repo":    cty.StringVal(computed.Repo),
		"url":     cty.StringVal(t.URL(t.Hash())),
		"output":  cty.StringVal(computed.Output),
		"version": cty.StringVal(t.Package.Version),
	})
}

// ComputedAttrs returns a pointer to computed attributes from the state store.
// If attributes are not in the state store it will create a new pointer and insert it into the state store.
func (t Target) ComputedAttrs() *ComputedAttrs {
	if attrs, ok := t.GetStateAttrs().(*ComputedAttrs); ok {
		return attrs
	}

	attrs := &ComputedAttrs{}

	t.SetStateAttrs(attrs)
	return attrs
}

// PreBuild a lifecycle hook for calculating state before the build
func (t Target) PreBuild() error {
	return hclutils.DecodeExpressions(&t, t.ComputedAttrs(), base.CreateEvalContext(base.EvalContextOptions{
		CurrentTarget:     t,
		Package:           *t.Package,
		TargetLookupTable: t.Workspace.TargetLUT,
		Workspace:         *t.Workspace,
	}))
}

func (t Target) createInMemoryContext(afs afero.Fs, tarFiles []*fs.TarFile, dump bool) error {
	archive, err := afs.OpenFile(ARK_CONTEXT_FILE, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer func() {
		_ = archive.Close()
	}()

	// Exclude raw Dockerfiles from the archive as we generate and inject it at runtime
	deps, _ := t.FileDeps()
	if archiveErr := deps.GzipArchive(t.Workspace.Dir, archive, func(file string) bool {
		return strings.HasSuffix(file, "Dockerfile") == false
	}, fs.InjectTarFiles(tarFiles)); archiveErr != nil {
		return archiveErr
	}

	if dump {
		if dumpErr := t.dumpContextToArtifact(afs); dumpErr != nil {
			return dumpErr
		}
	}

	return nil
}

func (t Target) dumpContextToArtifact(afs afero.Fs) error {
	err := t.MkArtifactsDir()
	if err != nil {
		return err
	}

	dumpDest, err := os.OpenFile(filepath.Join(t.ArtifactsDir(), ARK_CONTEXT_FILE), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer func() {
		_ = dumpDest.Close()
	}()

	dumpSource, err := afs.OpenFile(ARK_CONTEXT_FILE, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer func() {
		_ = dumpSource.Close()
	}()

	if _, cpErr := io.Copy(dumpDest, dumpSource); cpErr != nil {
		return cpErr
	}
	return nil
}

func (t Target) pullSecrets(secrets map[string]string) ([]secretsprovider.FileSource, error) {
	var arkSecretPaths []secretsprovider.FileSource

	for secretID, vaultPath := range secrets {
		contextualVaultPath := filepath.Join("secret/data", vaultPath)
		arkSecretPathDir := filepath.Join(t.Workspace.ArkDir(), "secrets", contextualVaultPath)
		arkSecretPathFile := filepath.Join(arkSecretPathDir, "data.json")
		if err := os.MkdirAll(arkSecretPathDir, 0700); err != nil {
			return arkSecretPaths, errors.Wrap(err, "unable to create directory for secret pull")
		}
		vaultData, err := t.Workspace.Vault.Logical().Read(contextualVaultPath)
		if res, ok := err.(*vault.ResponseError); ok {
			switch res.StatusCode {
			case http.StatusBadRequest:
				return arkSecretPaths, errors.Wrap(err, "a Vault token was not available at $HOME/.vault-token; please login to the Vault cli")
			case http.StatusForbidden:
				return arkSecretPaths, errors.Wrap(err, "the provided token was not able to access this secret path; please try re-logging into Vault")
			}
		}
		if err != nil {
			return arkSecretPaths, errors.Wrapf(err, "unable to read secret from path: %s", vaultPath)
		}

		vaultDataBytes, err := json.MarshalIndent(vaultData.Data["data"], "", " ")
		if err != nil {
			return arkSecretPaths, errors.Wrap(err, "unable to marshal Vault data to JSON")
		}
		if writeSecretFileErr := ioutil.WriteFile(arkSecretPathFile, vaultDataBytes, 0600); writeSecretFileErr != nil {
			return arkSecretPaths, errors.Wrapf(writeSecretFileErr, "unable to write secret to ark store path: %s", arkSecretPathFile)
		}
		arkSecretPaths = append(arkSecretPaths, secretsprovider.FileSource{
			ID:       secretID,
			FilePath: arkSecretPathFile,
		})
	}
	return arkSecretPaths, nil
}

// Build constructs a Docker image from the information provided in the docker_image target
func (t Target) Build() error {
	attrs := t.ComputedAttrs()
	memFS := afero.NewMemMapFs()

	if attrs.BuildArgs == nil {
		attrs.BuildArgs = make(map[string]*string)
	}

	if attrs.CacheInline {
		attrs.BuildArgs["BUILDKIT_INLINE_CACHE"] = &BUILDKIT_INLINE_CACHE
	}

	tarFiles := []*fs.TarFile{
		{
			Name: "Dockerfile",
			Body: []byte(t.generateDockerfile()),
			Mode: 0600,
		},
	}

	var secretSpecs []secretsprovider.FileSource
	if attrs.Secrets != nil {
		returnedSpecs, err := t.pullSecrets(attrs.Secrets)
		if err != nil {
			return err
		}
		secretSpecs = returnedSpecs
	}

	if err := t.createInMemoryContext(memFS, tarFiles, attrs.DumpContext); err != nil {
		return err
	}

	dockerContext, err := memFS.OpenFile(ARK_CONTEXT_FILE, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}

	defer func() {
		_ = dockerContext.Close()
	}()

	eg, ctx := errgroup.WithContext(appCTX)
	sess, err := t.Workspace.Docker.StartBuildkitSession(eg, ctx, secretSpecs, utils.UUIDV4())
	if err != nil {
		return err
	}

	var imageOutputs []types.ImageBuildOutput

	if attrs.Output != "" {
		sess.Allow(filesync.NewFSSyncTargetDir(attrs.Output))
		imageOutputs = []types.ImageBuildOutput{
			{Type: "local",
				Attrs: map[string]string{
					"dest": attrs.Output,
				}},
		}
	}

	eg.Go(func() error {
		defer func() {
			_ = sess.Close()
		}()

		return t.Workspace.Docker.Build(eg, ctx, dockerContext, types.ImageBuildOptions{
			Tags:           t.URLsFromTags(),
			SuppressOutput: false,
			NoCache:        false,
			Remove:         false,
			ForceRemove:    false,
			PullParent:     false,
			Dockerfile:     "Dockerfile",
			BuildArgs:      attrs.BuildArgs,
			Ulimits:        nil,
			AuthConfigs:    nil,
			Squash:         false,
			CacheFrom:      attrs.CacheFrom,
			SecurityOpt:    nil,
			Target:         attrs.Target,
			Outputs:        imageOutputs,
			SessionID:      sess.ID(),
			BuildID:        stringid.GenerateRandomID(),
			Version:        types.BuilderBuildKit,
		})
	})

	return eg.Wait()
}

// URL constructs an image URL from the repo and a given tag
func (t Target) URL(tag string) string {
	return fmt.Sprintf("%s:%s", t.ComputedAttrs().Repo, tag)
}

// URLsFromTags creates a slice of image URLs
func (t Target) URLsFromTags() []string {
	imageTags := []string{
		t.URL(t.Hash()),
		t.URL(t.ShortHash()),
		t.URL("latest"),
	}

	for _, tag := range t.ComputedAttrs().Tags {
		imageTags = append(imageTags, t.URL(tag))
	}
	return imageTags
}

func (t Target) generateDockerfile() string {
	dockerFile := t.ComputedAttrs().Dockerfile

	// disabled entrypoint injection at the workspace level
	if (t.Workspace.Config.Internal != nil && t.Workspace.Config.Internal.DisableEntrypointInjection) ||
		// disabled entrypoint injection at the workspace level
		t.ComputedAttrs().DisableEntrypointInjection {
		log.Debugf("entrypoint injection was disabled for %s", t.Address())
		return dockerFile
	}

	imageURL := "gcr.io/[insert-google-project]/domain/ark-entrypoint:7a227a8"
	fromLine := fmt.Sprintf("FROM %s as ark-entrypoint", imageURL)
	copyLine := fmt.Sprintf(arkEPCopyTemplate, t.Address(), t.Hash())
	lines := strings.Split(dockerFile, "\n")
	firstLine := strings.TrimSpace(lines[0])

	if strings.HasPrefix(firstLine, "#") {
		lines[0] = strings.Join([]string{firstLine, fromLine}, "\n")
	} else {
		lines = append([]string{fromLine}, lines...)
	}
	lines = append(lines, copyLine)

	return strings.Join(lines, "\n")
}

// CheckLocalBuildCache loads the build cache state and verifies that the hashes match
func (t Target) CheckLocalBuildCache() (bool, error) {
	if t.ComputedAttrs().Output == "" {
		log.Debugf("checking for local Docker image: %s", t.URL(t.Hash()))
		exists, err := t.Workspace.Docker.ImageExists(appCTX, t.URL(t.Hash()))
		if err != nil {
			return false, errors.Wrap(err, "CheckLocalBuildCache#ImageExists")
		}
		return exists, nil
	}
	return t.RawTarget.CheckLocalBuildCache()
}

// CheckRemoteCache determines whether a docker image for a given target exists in a remote repository
func (t Target) CheckRemoteCache() (bool, error) {
	if t.ComputedAttrs().Output == "" {
		exists, err := t.Workspace.Docker.RepoImageExists(appCTX, t.URL(t.Hash()))
		if err != nil {
			return false, errors.Wrap(err, "CheckRemoteCache#RepoImageExists")
		}
		return exists, nil
	}
	return t.RawTarget.CheckRemoteCache()
}

// PullRemoteCache pulls a docker image from a remote repository that matches a target's state hash
func (t Target) PullRemoteCache() error {
	if t.ComputedAttrs().Output == "" {
		return t.Workspace.Docker.PullImage(appCTX, t.URL(t.Hash()))
	}
	return t.RawTarget.PullRemoteCache()
}

// PushRemoteCache pushes a state artifact to remote storage and pushes a docker image to a remote repository
func (t Target) PushRemoteCache() error {
	if t.ComputedAttrs().Output == "" {
		for _, imageURL := range t.URLsFromTags() {
			if err := t.Workspace.Docker.PushImage(appCTX, imageURL); err != nil {
				return errors.Wrap(err, "an error occurred while pushing the images")
			}
		}
		return nil
	}
	return t.RawTarget.PushRemoteCache()
}
