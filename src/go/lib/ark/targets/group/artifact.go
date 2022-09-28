package group

import (
	"context"

	"github.com/myfintech/ark/src/go/lib/ark"
)

// Artifact is the result of a successful actions.Group
type Artifact struct {
	ark.RawArtifact `mapstructure:",squash"`
}

// Cacheable always returns false because the targets the group target is running should be assessed for their cache state individually
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

// Pull does not do anything
func (a Artifact) Pull(_ context.Context) error {
	return nil
}
