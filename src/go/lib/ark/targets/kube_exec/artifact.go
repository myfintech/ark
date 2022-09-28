package kube_exec

import (
	"context"

	"github.com/myfintech/ark/src/go/lib/ark"
)

// Artifact the result of a successful kube_exec.Produce() call
type Artifact struct {
	ark.RawArtifact `mapstructure:",squash"`
}

// Cacheable always returns false as an exec should run regardless of state
func (a Artifact) Cacheable() bool {
	return false
}

// RemotelyCached is an unused function as the target is not cacheable
func (a Artifact) RemotelyCached(_ context.Context) (bool, error) {
	return false, nil
}

// LocallyCached is an unused function as the target is not cacheable
func (a Artifact) LocallyCached(_ context.Context) (bool, error) {
	return false, nil
}

// Push is an unused function as the target is not cacheable
func (a Artifact) Push(_ context.Context) error {
	return nil
}

// Pull is an unused function as the target is not cacheable
func (a Artifact) Pull(_ context.Context) error {
	return nil
}
