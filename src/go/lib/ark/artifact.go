package ark

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/myfintech/ark/src/go/lib/ark/workspace"

	"github.com/myfintech/ark/src/go/lib/fs"

	"github.com/myfintech/ark/src/go/lib/utils/cloudutils"

	"github.com/pkg/errors"
)

// RawArtifact represents the core representation of any artifact
type RawArtifact struct {
	Key                string                 `json:"key" mapstructure:"key"`
	Hash               string                 `json:"hash" mapstructure:"hash"`
	Type               string                 `json:"type" mapstructure:"type"`
	Attributes         map[string]interface{} `json:"attributes" mapstructure:"attributes,remain"`
	DependsOn          Ancestors              `json:"dependsOn" mapstructure:"dependsOn" hash:"-"`
	RemoteCacheBaseURL string                 `json:"remote_cache_base_url" mapstructure:"remote_cache_base_url"`
}

// Artifact is an interface that defines the behavior of a resource that can be stored and fetched from a content addressable store
type Artifact interface {
	Cacheable() bool
	RemotelyCached(ctx context.Context) (bool, error)
	LocallyCached(ctx context.Context) (bool, error)
	Push(ctx context.Context) error
	Pull(ctx context.Context) error
	MkCacheDir() (string, error)
	WriteState() error
}

// Key is a convenience struct to assist with directory creation
type Key struct {
	Path string
	Name string
}

// ParseKey takes an artifact key and marshals the contents into the Key struct
func ParseKey(key string) (Key, error) {
	keyComponents := strings.Split(key, ":")
	if len(keyComponents) != 2 {
		return Key{}, errors.New("the provided key is invalid and should be in the format of '<path>:<name>'")
	}

	return Key{
		Path: keyComponents[0],
		Name: keyComponents[1],
	}, nil
}

// UseWorkspaceConfig injects the blob storage base URL into the raw artifact struct
func (r *RawArtifact) UseWorkspaceConfig(config workspace.Config) {
	r.RemoteCacheBaseURL = config.RemoteCache.URL
}

// Cacheable always returns true as the default behavior should be to cache something
func (r RawArtifact) Cacheable() bool {
	return true
}

// RemotelyCached checks a remote blob store for the presence of an artifact
func (r RawArtifact) RemotelyCached(ctx context.Context) (bool, error) {
	if r.RemoteCacheBaseURL == "" {
		return false, errors.New("a remote cache URL has not been provided in the workspace configuration")
	}
	return cloudutils.BlobCheck(ctx, r.RemoteCacheBaseURL, fmt.Sprintf("%s.tar.gz", r.Hash))
}

// LocallyCached checks for the presence of a local artifact.json file
func (r RawArtifact) LocallyCached(_ context.Context) (bool, error) {
	localCacheDir, err := r.CacheDirPath()
	if err != nil {
		return false, err
	}

	if _, err = os.Stat(filepath.Join(localCacheDir, "artifact.json")); os.IsNotExist(err) {
		return false, err
	}

	return true, err
}

// Push uploads an artifact to a remote blob store
func (r RawArtifact) Push(ctx context.Context) error {
	if r.RemoteCacheBaseURL == "" {
		return errors.New("a remote cache URL has not been provided in the workspace configuration")
	}

	writer, cleanup, err := cloudutils.NewBlobWriter(ctx, r.RemoteCacheBaseURL, fmt.Sprintf("%s.tar.gz", r.Hash))
	defer func() {
		if cleanup != nil {
			cleanup()
		}
	}()
	if err != nil {
		return err
	}

	localCacheDir, err := r.CacheDirPath()
	if err != nil {
		return err
	}

	if err = fs.GzipTar(localCacheDir, writer); err != nil {
		return err
	}

	return nil
}

// Pull downloads an artifact from a remote blob store to an expected location on disk
func (r RawArtifact) Pull(ctx context.Context) error {
	if r.RemoteCacheBaseURL == "" {
		return errors.New("a remote cache URL has not been provided in the workspace configuration")
	}

	reader, cleanup, err := cloudutils.NewBlobReader(ctx, r.RemoteCacheBaseURL, fmt.Sprintf("%s.tar.gz", r.Hash))
	defer func() {
		if cleanup != nil {
			cleanup()
		}
	}()
	if err != nil {
		return err
	}

	localCacheDir, err := r.MkCacheDir()
	if err != nil {
		return err
	}
	if err = fs.GzipUntar(localCacheDir, reader); err != nil {
		return err
	}

	return nil
}

// CacheDirPath returns the constructed name of the local filesystem path for local caching of an artifact
func (r RawArtifact) CacheDirPath() (string, error) {
	key, err := ParseKey(r.Key)
	if err != nil {
		return "", err
	}

	cacheDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(cacheDir, "ark", "artifacts", key.Name, r.Hash), nil
}

// MkCacheDir creates a location on disk for storing an artifact locally
func (r RawArtifact) MkCacheDir() (string, error) {
	cacheDir, err := r.CacheDirPath()
	if err != nil {
		return "", err
	}

	if err = os.MkdirAll(cacheDir, 0755); err != nil {
		return "", err
	}

	return cacheDir, nil
}

// WriteState writes a json representation of the raw artifact to the local cache directory
func (r RawArtifact) WriteState() error {
	cacheDir, err := r.MkCacheDir()
	if err != nil {
		return err
	}

	jsonBytes, err := json.MarshalIndent(r, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(cacheDir, "artifact.json"), jsonBytes, 0644)
}

// ShortHash returns a 7 character version of the has
// returns an empty string if the hash is empty
func (r RawArtifact) ShortHash() string {
	if len(r.Hash) < 8 {
		return r.Hash
	}
	return r.Hash[0:7]
}
