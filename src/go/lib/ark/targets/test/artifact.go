package test

import (
	"context"

	"github.com/myfintech/ark/src/go/lib/ark"
)

// Artifact the result of a successful test.Produce() call
type Artifact struct {
	ark.RawArtifact `mapstructure:",squash"`
}

// Cacheable currently returns false as caching for a test is very nebulous and needs more thought
func (a Artifact) Cacheable() bool {
	return false
}

// RemotelyCached currently does nothing as caching is disabled
func (a Artifact) RemotelyCached(_ context.Context) (bool, error) {
	return false, nil
}

// LocallyCached currently does nothing as caching is disabled
func (a Artifact) LocallyCached(_ context.Context) (bool, error) {
	return false, nil
}

// Push currently does nothing as caching is disabled
func (a Artifact) Push(_ context.Context) error {
	return nil
}

// Pull currently does nothing as caching is disabled
func (a Artifact) Pull(_ context.Context) error {
	return nil
}
