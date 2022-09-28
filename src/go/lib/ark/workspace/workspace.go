package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/myfintech/ark/src/go/lib/kube"
	"github.com/pkg/errors"
)

type InformersConfig struct {
	ResyncPeriod time.Duration `json:"resyncPeriod"`
}

// KubernetesConfig holds data for applying k8s manifests to clusters
type KubernetesConfig struct {
	SafeContexts []string        `json:"safe_contexts"`
	Namespace    string          `json:"namespace"`
	Informers    InformersConfig `json:"infomers"`
}

// GetSafeContexts return a list of safe contexts from the workspace if any + ARK_K8S_SAFE_CONTEXTS system environment variable + kube.DefaultSafeContexts
func (k8s KubernetesConfig) GetSafeContexts() []string {
	var safeContexts []string
	safeContexts = append(safeContexts, kube.DefaultSafeContexts()...)
	safeContexts = append(safeContexts, strings.Split(os.Getenv("ARK_K8S_SAFE_CONTEXTS"), ",")...)
	if k8s.SafeContexts != nil {
		safeContexts = append(safeContexts, k8s.SafeContexts...)
	}
	return safeContexts
}

// FileSystemConfig configures the workspace file system observer
type FileSystemConfig struct {
	Ignore []string `json:"ignore"`
}

// RemoteCacheConfig configures the workspace remote cache location
type RemoteCacheConfig struct {
	URL string `json:"url"`
}

// VaultConfig allows user to set Vault address that's not reliant on an env var
type VaultConfig struct {
	Address       string `json:"address"`
	EncryptionKey string `json:"encryption_key"`
}

// Plugin defines a docker image that can be used as a function
// the image must take in data from stdin and the output from the plugin is a string
type Plugin struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

// ControlPlaneConfig controls options injected into ark control plane components
type ControlPlaneConfig struct {
	OrgID        string `json:"org_id"`
	ProjectID    string `json:"project_id"`
	ApiURL       string `json:"api_url"`
	EventSinkURL string `json:"event_sink_url"`
	LogSinkURL   string `json:"log_sink_url"`
}

// UserConfig used to authenticate users of the ark control plane
type UserConfig struct {
	Token string `json:"address"`
}

// InternalConfig an object which describes arks internal configuration options
type InternalConfig struct {
	DisableEntrypointInjection bool `json:"disable_entrypoint_injection"`
}

// Config holds data for configuring a workspace
type Config struct {
	file                 string
	K8s                  KubernetesConfig   `json:"kubernetes"`
	Vault                VaultConfig        `json:"vault"`
	FileSystem           FileSystemConfig   `json:"file_system"`
	RemoteCache          RemoteCacheConfig  `json:"remote_cache"`
	Plugins              []Plugin           `json:"plugins"`
	ControlPlane         ControlPlaneConfig `json:"control_plane"`
	User                 UserConfig         `json:"user"`
	Internal             InternalConfig     `json:"internal"`
	VersionCheckDisabled bool               `json:"disable_version_check"`
}

// DetermineRoot recursively searches the current directory and all its parents for .ark/settings.json
func DetermineRoot(startingDir string) (string, error) {
	configRoot, err := filepath.Abs(startingDir)
	if err != nil {
		return configRoot, err
	}

	_, err = os.Stat(configFilePath(configRoot))

	// break recursion because we've reached the root file system
	if os.IsNotExist(err) && configRoot == "/" {
		return configRoot, errors.New(
			".ark/settings.json not found after traversing all parent directories",
		)
	}

	// recurse into the parent dir
	if os.IsNotExist(err) {
		return DetermineRoot(filepath.Join(configRoot, ".."))
	}

	return configRoot, nil
}

// DetermineRootFromCWD determine the workspace root from the current working directory
func DetermineRootFromCWD() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return DetermineRoot(cwd)
}

// LoadConfig reads the bytes out of json settings and unmarshalls them into a Config object
func LoadConfig(startingDir string) (*Config, error) {
	c := new(Config)

	configRoot, err := DetermineRoot(startingDir)
	if err != nil {
		return nil, err
	}

	configBytes, err := os.ReadFile(configFilePath(configRoot))
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(configBytes, c); err != nil {
		return nil, err
	}

	c.file = configFilePath(configRoot)

	return c, nil
}

func configFilePath(configDir string) string {
	return filepath.Join(configDir, ".ark/settings.json")
}

// LoadConfigFromCWD reads the bytes out of json settings and unmarshalls them into a Config object from the CWD
func LoadConfigFromCWD() (*Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return LoadConfig(cwd)
}

func (c Config) File() string {
	return c.file
}

func (c Config) Root() string {
	return filepath.Clean(filepath.Join(c.Dir(), ".."))
}

func (c Config) Dir() string {
	return filepath.Dir(c.file)
}
