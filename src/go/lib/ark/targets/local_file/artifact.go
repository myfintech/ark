package local_file

import (
	"context"

	"github.com/myfintech/ark/src/go/lib/ark"
)

// Artifact the result of a successful actions.LocalFile
type Artifact struct {
	ark.RawArtifact  `mapstructure:",squash"`
	RenderedFilePath string `json:"renderedFilePath" mapstructure:"renderedFilePath"`
}

// Cacheable always returns false because the contents of a generated file might change
// depending on the configuration provided to the target at runtime
func (a Artifact) Cacheable() bool {
	return false
}

// RemotelyCached always returns false because the target is not cacheable
func (a Artifact) RemotelyCached(_ context.Context) (bool, error) {
	return false, nil
}

// LocallyCached always returns false because the target is not cacheable
func (a Artifact) LocallyCached(_ context.Context) (bool, error) {
	return false, nil
}

// Push does not do anything
func (a Artifact) Push(_ context.Context) error {
	return nil
}

// Pull does nothing
func (a Artifact) Pull(_ context.Context) error {
	return nil
}
