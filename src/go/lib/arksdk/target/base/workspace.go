package base

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/myfintech/ark/src/go/lib/ark/kv"
	"github.com/myfintech/ark/src/go/lib/xdgbase"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/moby/buildkit/util/appcontext"
	"github.com/pkg/errors"
	"github.com/reactivex/rxgo/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zclconf/go-cty/cty"

	"github.com/myfintech/ark/src/go/lib/container"
	"github.com/myfintech/ark/src/go/lib/dag"
	"github.com/myfintech/ark/src/go/lib/fs"
	"github.com/myfintech/ark/src/go/lib/fs/observer"
	"github.com/myfintech/ark/src/go/lib/git/gitignore"
	"github.com/myfintech/ark/src/go/lib/hclutils"
	"github.com/myfintech/ark/src/go/lib/jsonnetutils"
	"github.com/myfintech/ark/src/go/lib/kube"
	"github.com/myfintech/ark/src/go/lib/kube/portbinder"
	"github.com/myfintech/ark/src/go/lib/log"
	"github.com/myfintech/ark/src/go/lib/utils"
	"github.com/myfintech/ark/src/go/lib/watchman"

	vault "github.com/hashicorp/vault/api"
	jsonnetlib "github.com/myfintech/ark/src/jsonnet/lib"
)

var arkHome string

func init() {
	dir, err := xdgbase.Dir("ark", xdgbase.DataSuffix)
	if err != nil {
		log.Fatalf("failed to initialize data home %v", err)
	}
	arkHome = dir
}

var ctxLog = log.WithFields(log.Fields{
	"prefix": "arksdk",
})

// ArtifactsConfig holds data for the storage of artifacts
type ArtifactsConfig struct {
	StorageBaseURL string `hcl:"storage_base_url,attr"`
}

// KubernetesConfig holds data for applying k8s manifests to clusters
type KubernetesConfig struct {
	SafeContexts []string `hcl:"safe_contexts,attr"`
}

// FileSystemConfig configures the workspace file system observer
type FileSystemConfig struct {
	Ignore []string `hcl:"ignore,attr"`
}

// VaultConfig allows user to set Vault address that's not reliant on an env var
type VaultConfig struct {
	Address       string `hcl:"address,attr"`
	EncryptionKey string `hcl:"encryption_key,optional"`
}

// JsonnetConfig allows a user to configure overrides for internal systems
type JsonnetConfig struct {
	Library []string `hcl:"library,optional"`
}

// Plugin defines a docker image that can be used as an HCL function
// the image must take in data from stdin and the output from the plugin is a string
type Plugin struct {
	Name  string `hcl:"name,label"`
	Image string `hcl:"image,attr"`
}

// ControlPlaneConfig controls options injected into ark control plane components
type ControlPlaneConfig struct {
	OrgID        string `hcl:"org_id,attr"`
	ProjectID    string `hcl:"project_id,attr"`
	ApiURL       string `hcl:"api_url,attr"`
	EventSinkURL string `hcl:"event_sink_url,attr"`
	LogSinkURL   string `hcl:"log_sink_url,attr"`
}

// UserConfig used to authenticate users of the ark control plane
type UserConfig struct {
	Token string `hcl:"address,attr"`
}

// InternalConfig an object which describes arks internal configuration options
type InternalConfig struct {
	DisableEntrypointInjection bool `hcl:"disable_entrypoint_injection,optional"`
}

// WorkspaceConfig holds data for configuring a workspace
type WorkspaceConfig struct {
	K8s                  *KubernetesConfig   `hcl:"kubernetes,block"`
	Vault                *VaultConfig        `hcl:"vault,block"`
	Artifacts            *ArtifactsConfig    `hcl:"artifacts,block"`
	FileSystem           *FileSystemConfig   `hcl:"file_system,block"`
	Jsonnet              *JsonnetConfig      `hcl:"jsonnet,block"`
	Plugins              []Plugin            `hcl:"plugin,block"`
	ControlPlane         *ControlPlaneConfig `hcl:"control_plane,block"`
	User                 *UserConfig         `hcl:"user,block"`
	Internal             *InternalConfig     `hcl:"internal,block"`
	VersionCheckDisabled *bool               `hcl:"disable_version_check,attr"`
}

// K8sConfig performs a nil check on w.K8s and returns a pointer to a k8s config
func (w WorkspaceConfig) K8sConfig() *KubernetesConfig {
	if w.K8s != nil {
		return w.K8s
	}
	return &KubernetesConfig{}
}

// Workspace the workspace configuration for ark
type Workspace struct {
	Dir  string
	File string
	HCL  hcl.Body `json:"-"`

	TargetWatch       bool
	TargetGraph       Graph
	RegisteredTargets Targets
	TargetLUT         LookupTable
	Observer          *observer.Observer
	Config            WorkspaceConfig
	Context           context.Context

	K8s       kube.Client
	Vault     *vault.Client
	Docker    *container.Docker
	KVStorage kv.Storage

	Cmd                   *cobra.Command
	Args                  []string
	PassableArgs          []string
	DefaultJsonnetLibrary []string

	PortBinderCommands       portbinder.CommandChannel
	ReadyPortCommands        portbinder.CommandChannel
	ConfigurationEnvironment string

	VaultClientFactory VaultClientFactory
}

// JsonnetLibrary returns a list of paths to jsonnet libraries
func (w *Workspace) JsonnetLibrary(ext []string) []string {
	var workspaceLibraries []string
	if w.Config.Jsonnet != nil {
		workspaceLibraries = w.Config.Jsonnet.Library
	}
	return jsonnetutils.BuildLibrary(w.Dir, w.DefaultJsonnetLibrary, workspaceLibraries, ext)
}

// DetermineRoot recursively searches the current directory and all its parents for a WORKSPACE.hcl file
func (w *Workspace) DetermineRoot(startingDir string) error {
	workspaceRoot, err := filepath.Abs(startingDir)
	if err != nil {
		return err
	}

	workspaceFile := filepath.Join(workspaceRoot, "WORKSPACE.hcl")

	_, err = os.Stat(workspaceFile)

	// break recursion because we've reached the root file system
	// FIXME: this will only work on Linux/MacOS (not really important... for now)
	// There's probably a variable or function that lives in the os package we should leverage
	if os.IsNotExist(err) && workspaceRoot == "/" {
		return errors.New("no WORKSPACE.hcl found after traversing all parent directories")
	}

	// recurse into the parent dir
	if os.IsNotExist(err) {
		return w.DetermineRoot(filepath.Join(workspaceRoot, ".."))
	}

	patterns, err := gitignore.LoadRepoPatterns(workspaceRoot)

	if err != nil {
		return errors.Wrap(err, "failed to load .gitignore")
	}

	// bingo!
	w.Dir = workspaceRoot
	w.File = workspaceFile

	socket, _ := watchman.GetSocketName()
	nativeModeEnabled := socket == "" // no watchman socket available in this environment
	ctxLog.WithField("native-mode-enabled", nativeModeEnabled).Debug("fs observer")
	w.Observer = observer.NewObserver(nativeModeEnabled, w.TargetWatch, w.Dir, []string{}, gitignore.NewMatcher(patterns), logz.NoOpLogger{}, nil)
	return nil
}

// DetermineRootFromCWD determine the workspace root from the current working directory
func (w *Workspace) DetermineRootFromCWD() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	return w.DetermineRoot(cwd)
}

// DirBase returns the basedir of the workspace instead of the absolute path
func (w *Workspace) DirBase() string {
	return filepath.Base(w.Dir)
}

// DecodeFile decodes the workspace file into the workspace configuration struct
func (w *Workspace) DecodeFile(eval *hcl.EvalContext) hcl.Diagnostics {
	hclFile, diag := hclutils.FileFromPath(w.File)
	if diag != nil && diag.HasErrors() {
		return diag
	}
	if hclFile != nil {
		w.HCL = hclFile.Body
	}
	if diag = gohcl.DecodeBody(w.HCL, eval, &w.Config); diag != nil && diag.HasErrors() {
		return diag
	}

	w.Config.User = &UserConfig{}
	userToken, err := ioutil.ReadFile("$HOME/.config/ark/token")
	if err != nil {
		log.Debug(err)
	} else {
		w.Config.User.Token = string(userToken)
	}

	if w.Config.FileSystem != nil {
		w.Observer.Ignore = w.Config.FileSystem.Ignore
	}

	if w.Config.VersionCheckDisabled != nil {
		viper.Set("skip_version_check", *w.Config.VersionCheckDisabled)
	}
	return nil
}

// InitKubeClient initializes a new kubernetes client set
func (w *Workspace) InitKubeClient(namespace string) error {
	if w.Config.K8s != nil {
		w.K8s = kube.Init(nil)
		safe := false
		currentContext, err := w.K8s.CurrentContext()
		if err != nil {
			return err
		}
		if currentContext == "" {
			return nil
		}
		for _, safeContext := range w.K8sSafeContexts() {
			if safe = safeContext == currentContext; safe {
				break
			}
		}
		if !safe {
			return errors.Errorf(`
				Your current k8s context(%s) is unsafe, please change it or
				update your WORKSPACE.hcl kubernetes.safe_contexts to include this context
			`, currentContext)
		}
		if namespace != "" {
			w.K8s.NamespaceOverride = namespace
		}
	}

	return nil

}

// K8sSafeContexts return a slice of k8s safe contexts
func (w *Workspace) K8sSafeContexts() []string {
	var safeContexts []string
	safeContexts = append(safeContexts, strings.Split(os.Getenv("ARK_K8S_SAFE_CONTEXTS"), ",")...)
	if w.Config.K8s != nil {
		safeContexts = append(safeContexts, w.Config.K8s.SafeContexts...)
	}
	return safeContexts
}

// InitVaultClient initializes a new Vault client
func (w *Workspace) InitVaultClient() error {
	vaultAddr := utils.EnvLookup("VAULT_ADDR", "")
	if w.Config.Vault != nil {
		vaultAddr = w.Config.Vault.Address
	}
	ctxLog.Debugf("vault address used for client is set to %s", vaultAddr)
	client, err := vault.NewClient(&vault.Config{
		Address:    vaultAddr,
		HttpClient: nil,
	})
	if err != nil {
		return err
	}
	w.Vault = client

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "unable to get user home directory path")
	}

	tokenFile := filepath.Join(homeDir, ".vault-token")
	vaultTokenBytes, err := ioutil.ReadFile(tokenFile)
	if os.IsNotExist(err) {
		ctxLog.Warnf("vault token file does not exist: %s", tokenFile)
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "unable to read Vault token from file: %s", tokenFile)
	}

	w.Vault.SetToken(string(vaultTokenBytes))

	if w.Config.Vault != nil {
		w.KVStorage = &kv.VaultStorage{
			Client:        w.Vault,
			FSBasePath:    filepath.Join(w.Dir, ".ark/kv"),
			EncryptionKey: w.Config.Vault.EncryptionKey,
		}
	}
	return nil
}

// LoadBuildFiles walks the workspace root, and it's children for BUILD.hcl files
func (w *Workspace) LoadBuildFiles() ([]string, error) {
	var buildFiles []string

	return buildFiles, filepath.Walk(w.Dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		// if this file is under a sub workspace we've seen before we should ignore it
		// if we see a directory first check it for a WORKSPACE.hcl file
		// if there is one this is a sub workspace and should not be traversed
		if info.IsDir() && path != w.Dir {
			if _, wspErr := os.Stat(filepath.Join(path, "WORKSPACE.hcl")); wspErr == nil {
				return filepath.SkipDir
			}
		}

		if info.IsDir() && w.Observer.GitIgnoreMatcher.Match(fs.Split(path), info.IsDir()) {
			ctxLog.Tracef("build file walk skipping %s", path)
			return filepath.SkipDir
		}

		// if this is a BUILD.hcl file we should collect it
		if strings.Contains(path, "BUILD.hcl") {
			buildFiles = append(buildFiles, path)
			return nil
		}

		return nil
	})
}

// DecodeBuildFiles loads build files from the workspace and decodes them into build file structs.
func (w *Workspace) DecodeBuildFiles() ([]BuildFile, error) {
	var buildFiles []BuildFile
	buildFilePaths, err := w.LoadBuildFiles()
	if err != nil {
		return buildFiles, err
	}

	if len(buildFilePaths) == 0 {
		return buildFiles, errors.Errorf("no BUILD.hcl files found located in %s", w.Dir)
	}
	for _, buildFilePath := range buildFilePaths {
		// parse build files as HCL
		hclFile, diag := hclutils.FileFromPath(buildFilePath)
		if diag != nil && diag.HasErrors() {
			return buildFiles, diag
		}
		buildFiles = append(buildFiles, BuildFile{
			HCL:  hclFile,
			Path: buildFilePath,
		})
	}
	return buildFiles, nil
}

// Attributes of this workspace, intended to be used in hcl.EvalContext.
func (w Workspace) Attributes() map[string]interface{} {
	return map[string]interface{}{
		"path": w.Dir,
		"config": map[string]interface{}{
			"kubernetes": map[string]interface{}{
				"safe_contexts": w.Config.K8sConfig().SafeContexts,
			},
		},
	}
}

// AttributesToCty returns the workspace attributes as cty values
func (w Workspace) AttributesToCty() map[string]cty.Value {
	return hclutils.MapStringInterfaceToCty(w.Attributes())
}

// ArkDir returns the /ark directory relative to the workspace root
func (w *Workspace) ArkDir() string {
	return arkHome
}

// ArtifactsDir returns the .ark/artifacts directory relative to the workspace root
func (w *Workspace) ArtifactsDir() string {
	return filepath.Join(w.ArkDir(), "artifacts")
}

// Clean removes generated files from the /ark/artifacts directory
// this is used to reset the local state and artifacts cache
func (w Workspace) Clean() error {
	return os.RemoveAll(w.ArtifactsDir())
}

// NewWorkspace creates a workspace with defaults
func NewWorkspace() *Workspace {
	return &Workspace{
		Dir:     "",
		File:    "",
		HCL:     nil,
		Context: appcontext.Context(),

		TargetGraph:       Graph{},
		RegisteredTargets: Targets{},
		TargetLUT:         LookupTable{},
		Observer: &observer.Observer{
			Logger: logz.NoOpLogger{},
		},
		Config:                WorkspaceConfig{},
		DefaultJsonnetLibrary: []string{jsonnetlib.LibPath},
		PortBinderCommands:    make(portbinder.CommandChannel, 1000),
		ReadyPortCommands:     make(portbinder.CommandChannel, 1000),
	}
}

// BuildFile the parsed hcl file and location on the file system
type BuildFile struct {
	HCL  *hcl.File
	Path string
}

/*
LoadTargets
- loads all build files in the workspace
- attempts to decode targets into their respective types
- loads targets into the graph, and the look up table
- iterates over all loaded targets to build graph edges
- validates the graph for cycles and errors
*/
func (w *Workspace) LoadTargets(buildFiles []BuildFile) error {
	// iterate over build files
	for _, buildFile := range buildFiles {
		log.Debugf("loaded build file %s", buildFile.Path)

		// decode HCL files into rawTargets (decode pass 1)
		rawTargets, diag := DecodeRawTargetsFromHCLFile(DecodeRawTargetOpts{
			HclFile:          buildFile.HCL,
			Filename:         buildFile.Path,
			DecodedBuildFile: &DecodedBuildFile{},
			Workspace:        w,
			HclCtx: &hcl.EvalContext{
				Variables: map[string]cty.Value{
					"workspace": cty.ObjectVal(w.AttributesToCty()),
				},
			},
		})
		if diag != nil && diag.HasErrors() {
			return diag
		}

		log.Debugf("decoded %d raw targets", len(rawTargets))

		// decode rawTargets into their typed target structs (decode pass2)
		targets, diag := w.RegisteredTargets.MapRawTargets(rawTargets)
		if diag != nil && diag.HasErrors() {
			return diag
		}

		// load targets into the graph, and the look up table
		for _, target := range targets {
			if addErr := w.TargetLUT.Add(target); addErr != nil {
				return errors.Wrapf(addErr, "unable to load target '%s' into workspace lookup table", target.Address())
			}
		}
	}

	// after we've hydrated the graph and lookup table
	// build graph edges between the targets and their dependencies
	graph, err := w.TargetLUT.BuildGraph()
	if err != nil {
		return err
	}

	w.TargetGraph = *graph

	// validate the graph has no cycles and no basic errors
	if err = w.TargetGraph.Validate(); err != nil {
		return err
	}

	w.TargetGraph.TransitiveReduction()

	for _, addressable := range w.TargetLUT {
		if target, ok := addressable.(Target); ok {
			if err = target.EvaluateHCLExpressions(); err != nil {
				return err
			}
		}
	}

	cn, err := w.Observer.Reindex()
	if err != nil {
		return err
	}

	ctxLog.Tracef("%d files found", len(cn.Files))

	return nil
}

// WatchForChanges creates a fs observer subscription to the workspace observing changes to the root target
func (w *Workspace) WatchForChanges(filter rxgo.Predicate, onChange rxgo.Func) rxgo.Observable {
	return w.Observer.FileSystemStream.Filter(filter).Map(onChange)
}

// FilterChangeNotificationsByTarget is an rxgo.Predicate to filter a stream by workspace target dependencies
func FilterChangeNotificationsByTarget(graph *Graph, root dag.Vertex) rxgo.Predicate {
	return func(i interface{}) bool {
		n := i.(*observer.ChangeNotification)
		for _, vertex := range graph.TopologicalSort(root) {
			if addressable, isAddressable := vertex.(Addressable); isAddressable {
				if _, matched := n.Matched[addressable.Address()]; matched {
					log.Infof("%d matched", len(n.Matched))
					log.Infof("%s detected change", addressable.Address())
					return matched
				}
			}
		}
		return false
	}
}

// ReRunOnChange returns an rxgo.Func that lists the changes made against targets and then executes the provided callback
func ReRunOnChange(callback func() error) rxgo.Func {
	return func(_ context.Context, i interface{}) (interface{}, error) {
		n := i.(*observer.ChangeNotification)
		for key := range n.Matched {
			log.Debugf("target match %s", key)
		}
		return i, callback()
	}
}

// InitDockerClient creates a docker client on the workspace
func (w *Workspace) InitDockerClient() error {
	docker, err := container.NewDockerClient(container.DefaultDockerCLIOptions()...)
	if err != nil {
		return err
	}
	w.Docker = docker
	return nil
}

// GraphWalk walks the graph, calling your callback as each node is visited.
// This will walk nodes in parallel if it can. The resulting diagnostics
// contains problems from all graphs visited, in no particular order.
func (w *Workspace) GraphWalk(address string, walk GraphWalker) error {
	if address == "" {
		return w.TargetGraph.Walk(walk)
	}

	vertex, err := w.TargetLUT.LookupByAddress(address)
	if err != nil {
		return err
	}
	return w.TargetGraph.Isolate(vertex).Walk(walk)
}

// ExtractCliOptions gets all of the info passed in from the CLI
func (w *Workspace) ExtractCliOptions(cmd *cobra.Command, args []string) {
	w.Cmd = cmd
	w.Args = args
	if cmd.ArgsLenAtDash() != -1 {
		w.PassableArgs = args[cmd.ArgsLenAtDash():]
	}
}

// VaultClientFactory allows for the injection of a custom configured Vault client
type VaultClientFactory func(config *vault.Config) (*vault.Client, error)

// SetEnvironmentConstraint sets the environment constraint from the CLI flag for the purpose of configuration loading
func (w *Workspace) SetEnvironmentConstraint(envFlagValue string) {
	w.ConfigurationEnvironment = "local"
	if envFlagValue != "" {
		w.ConfigurationEnvironment = envFlagValue
	}
}
