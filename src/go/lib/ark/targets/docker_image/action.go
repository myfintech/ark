package docker_image

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/moby/buildkit/session/filesync"

	"github.com/myfintech/ark/src/go/lib/ark/kv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stringid"
	"github.com/moby/buildkit/session/secrets/secretsprovider"
	"github.com/spf13/afero"
	"golang.org/x/sync/errgroup"

	"github.com/myfintech/ark/src/go/lib/container"
	"github.com/myfintech/ark/src/go/lib/fs"
	"github.com/myfintech/ark/src/go/lib/utils"
)

var (
	buildKitInlineCache = "1"
)

const (
	contextFile        = "context.tar.gz"
	entrypointImageURL = "gcr.io/[insert-google-project]/domain/ark-entrypoint:7a227a8"
	arkEPCopyTemplate  = `COPY --from=ark-entrypoint /ark-entrypoint-linux /usr/local/bin/ark-entrypoint
ENV ARK_TARGET_ADDRESS=%s
ENV ARK_TARGET_HASH=%s
`
)

// Action is the executor for building a docker image
type Action struct {
	Target   *Target
	Artifact *Artifact
	Client   container.Docker
	KVStore  kv.Storage
	Logger   logz.FieldLogger
}

var _ logz.Injector = &Action{}

// UseLogger injects a logger into the target's action
func (a *Action) UseLogger(logger logz.FieldLogger) {
	a.Logger = logger
}

// UseKVStorage allows the KV Storage client to be injected into the target's action
func (a *Action) UseKVStorage(client kv.Storage) {
	a.KVStore = client
}

// UseDockerClient sets this Action container.Docker client
func (a *Action) UseDockerClient(client container.Docker) {
	a.Client = client
}

func (a Action) generateDockerFile() string {
	if a.Target.DisableEntrypointInjection {
		return a.Target.Dockerfile
	}

	fromLine := fmt.Sprintf("FROM %s as ark-entrypoint", entrypointImageURL)
	copyLine := fmt.Sprintf(arkEPCopyTemplate, a.Artifact.Key, a.Artifact.Hash)
	lines := strings.Split(a.Target.Dockerfile, "\n")
	firstLine := strings.TrimSpace(lines[0])

	if strings.HasPrefix(firstLine, "#") {
		lines[0] = strings.Join([]string{firstLine, fromLine}, "\n")
	} else {
		lines = append([]string{fromLine}, lines...)
	}
	lines = append(lines, copyLine)

	return strings.Join(lines, "\n")
}

// Execute runs the action and produces a docker_image.Artifact
func (a Action) Execute(ctx context.Context) (err error) {
	memFS := afero.NewMemMapFs()

	if a.Target.BuildArgs == nil {
		a.Target.BuildArgs = make(map[string]*string)
	}

	if a.Target.CacheInline {
		a.Target.BuildArgs["BUILDKIT_INLINE_CACHE"] = &buildKitInlineCache
	}

	tarFiles := []*fs.TarFile{
		{
			Name: "Dockerfile",
			Body: []byte(a.generateDockerFile()),
			Mode: 0600,
		},
	}

	var secretSpecs []secretsprovider.FileSource
	if a.Target.Secrets != nil {
		returnedSpecs, decryptErr := a.createSecretSpecs(a.Target.Secrets)
		if decryptErr != nil {
			return decryptErr
		}
		secretSpecs = returnedSpecs
	}

	defer func() {
		_ = a.removeSecretSpecs(secretSpecs)
	}()

	archive, err := memFS.OpenFile(contextFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return
	}
	defer func() {
		_ = archive.Close()
	}()

	// FIXME(sourcec0de): Exclude raw Dockerfiles from the archive as we generate and inject it at runtime
	if err = fs.GzipTarFiles(a.Target.SourceFiles, a.Target.Realm, archive, fs.InjectTarFiles(tarFiles)); err != nil {
		return
	}

	dockerContext, err := memFS.OpenFile(contextFile, os.O_RDONLY, 0666)
	if err != nil {
		return
	}

	defer func() {
		_ = dockerContext.Close()
	}()

	eg, ec := errgroup.WithContext(ctx)
	sess, err := a.Client.StartBuildkitSession(eg, ec, secretSpecs, utils.UUIDV4())
	if err != nil {
		return
	}

	var imageOutputs []types.ImageBuildOutput

	if a.Target.Output != "" {
		cacheDir, dirErr := a.Artifact.MkCacheDir()
		if dirErr != nil {
			return dirErr
		}

		output := filepath.Join(cacheDir, a.Target.Output)

		sess.Allow(filesync.NewFSSyncTargetDir(output))
		imageOutputs = []types.ImageBuildOutput{
			{Type: "local",
				Attrs: map[string]string{
					"dest": output,
				}},
		}
	}

	eg.Go(func() error {
		defer func() {
			_ = sess.Close()
		}()

		return a.Client.Build(eg, ctx, dockerContext, types.ImageBuildOptions{
			Tags: []string{
				a.Artifact.URL,
				fmt.Sprintf("%s:%s", a.Target.Repo, "latest"),
			},
			SuppressOutput: false,
			NoCache:        false,
			Remove:         false,
			ForceRemove:    false,
			PullParent:     false,
			Dockerfile:     "Dockerfile",
			Ulimits:        nil,
			AuthConfigs:    nil,
			Squash:         false,
			BuildArgs:      a.Target.BuildArgs,
			CacheFrom:      a.Target.CacheFrom,
			SecurityOpt:    nil,
			Outputs:        imageOutputs,
			SessionID:      sess.ID(),
			BuildID:        stringid.GenerateRandomID(),
			Version:        types.BuilderBuildKit,
		})
	})

	return eg.Wait()
}

func (a *Action) createSecretSpecs(secrets []string) ([]secretsprovider.FileSource, error) {
	var arkSecretPaths []secretsprovider.FileSource

	for _, secret := range secrets {
		secretFile, err := a.KVStore.DecryptToFile(secret)
		if err != nil {
			return arkSecretPaths, err
		}

		arkSecretPaths = append(arkSecretPaths, secretsprovider.FileSource{
			ID:       secret,
			FilePath: secretFile,
		})
	}
	return arkSecretPaths, nil
}

func (a *Action) removeSecretSpecs(secretSpecs []secretsprovider.FileSource) error {
	for _, spec := range secretSpecs {
		if err := os.RemoveAll(spec.FilePath); err != nil {
			return err
		}
	}
	return nil
}
