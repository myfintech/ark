package base

import (
	"github.com/zclconf/go-cty/cty"

	"github.com/myfintech/ark/src/go/lib/fs/observer"
	"github.com/myfintech/ark/src/go/lib/state_store"
)

// Addressable an interface describing objects which can be addressed
type Addressable interface {
	// Address is the string representation of the targets address in the build graph (root.go_binary.tool)
	Address() string

	// Describe returns the description of the target defined in the BUILD.hcl file (this is optional, so it can be empty)
	Describe() string

	// String is an alias for Address
	String() string

	// Hashcode is an alias for Address used to satisfy the hashable interface
	Hashcode() interface{}

	// ListLabels returns a string slice of labels on a target
	ListLabels() []string
}

// Target represents the base interface for all targets which must also be Addressable
type Target interface {
	Addressable

	// Deps is a static list of ancestor target addresses declared in the BUILD.hcl file (this can be empty)
	Deps() []string

	// FileDeps returns an observer file match cache
	FileDeps() (*observer.FileMatchCache, bool)

	// State is a KV state store for caching data about the buildable target
	State() state_store.KVStore

	// PrevState is a secondary state store intended for comparing current state to previously known state
	PrevState() state_store.KVStore

	// LocalVars are a list of cty values used to keep HCL dry
	LocalVars() map[string]cty.Value

	// Attributes are a static list of cty values to be used in HCL EvaluationContexts
	Attributes() map[string]cty.Value

	// DirRelToWorkspace is the directory of target relative to its workspace
	// This is typically the location of the HCL file the target was declared in
	DirRelToWorkspace() string

	// EvaluateHCLExpressions evaluate source_files, include_patters, and exclude_patterns as expressions
	EvaluateHCLExpressions() error

	// PackageName returns the string value of package.Name
	PackageName() string

	// PackageVersion returns the string value of package.Version
	PackageVersion() string

	// PackageDir is the absolute path of the directory of a target
	PackageDir() string

	GetRawTarget() *RawTarget
}

// Hashable an interface describing objects that implement hashing methods
type Hashable interface {
	// Hash a full hex string sha256sum of the buildable
	// This can be empty if the hash has not been computed and saved in the state store
	Hash() string

	// ShortHash a 7 character hex string sha256sum of the buildable
	// This can be empty if the hash has not been computed and saved in the state store
	ShortHash() string

	// HashableAttributes returns a map of attributes that are hashed for the purpose of tracking build configuration changes for a target
	HashableAttributes() map[string]interface{}
}

// Buildable represents a an object that can be built and produces artifacts
type Buildable interface {
	Target
	Hashable

	// PreBuild a lifecycle hook that should ALWAYS be called before Build
	// This is a stateful function that causes target state to be updated prior to a build execution
	PreBuild() error

	// Build a function that executes a build for a given target.
	// If this function succeeds then the build was successful and should have resulted in an artifact.
	Build() error

	// Artifacts is a list of declared artifacts names
	Artifacts() map[string]string

	// ArtifactsDir is the absolute path to the /ark/artifacts directory of this target
	ArtifactsDir() string

	// MkArtifactsDir recursively ensures the ArtifactsDir exists
	MkArtifactsDir() error
}

// Cacheable represents a union type of the Hashable, Buildable, and methods that allow its artifacts to be cached locally and remotely
type Cacheable interface {
	// CacheEnabled allows for overriding the default caching behavior on a per-target basis
	// TODO: Split cacheable functions from raw_target. This is a temporary hack.
	CacheEnabled() bool

	// StateFilePath is the path in the artifacts directory to the state.json file for this target
	StateFilePath() string

	// CheckLocalBuildCache checks if the local cache contains a copy of the buildable artifacts
	CheckLocalBuildCache() (bool, error)

	// LoadLocalBuildCacheState returns the contents of the state.json file at StateFilePath()
	LoadLocalBuildCacheState() (BuildCacheState, error)

	// SaveLocalBuildCacheState saves the current state to state.json file at StateFilePath()
	SaveLocalBuildCacheState() error

	// CheckRemoteCache checks if the remote cache contains a copy of the buildable artifacts
	CheckRemoteCache() (bool, error)

	// PullRemoteCache pulls a copy of the buildable artifacts from the remote cache
	PullRemoteCache() error

	// PushRemoteCache pushes a copy of the buildable artifacts to the remote cache
	PushRemoteCache() error
}

// BuildCacheState provides the structure for the state.json file
type BuildCacheState struct {
	Name string
	Type string
	Hash string
}
