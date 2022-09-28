package nix

import (
	"context"
	"encoding/json"
	"os/exec"

	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/log"
)

// Artifact is the result of a successful nix.Produce() call
type Artifact struct {
	ark.RawArtifact `mapstructure:",squash"`
	Packages        []string `json:"packages" mapstructure:"packages"`
}

// NxPkg is a container for unmarshalled nix query responses
type NxPkg struct {
	Name    string `json:"name"`
	Pname   string `json:"pname"`
	Version string `json:"version"`
	System  string `json:"system"`
}

// Cacheable returns true because a local nix store might have the package(s) installed, but remote caching checks will always return false
func (a Artifact) Cacheable() bool {
	return true
}

// RemotelyCached always returns false because if a package isn't locally cached, we should just install it using nix
func (a Artifact) RemotelyCached(_ context.Context) (bool, error) {
	log.Debug("The nix target does not use a remote cache, so it will always return 'false' with no error on a remote cache check")
	return false, nil
}

// LocallyCached queries the local nix store for the packages expressed in the target definition
func (a Artifact) LocallyCached(_ context.Context) (bool, error) {
	queryResponse := make(map[string]NxPkg)

	out, err := exec.Command("nix-env", "--query", "--installed", "--json").Output()
	if err != nil {
		return false, err
	}

	if err = json.Unmarshal(out, &queryResponse); err != nil {
		return false, err
	}

	availablePkgs := make(map[string]bool)

	for _, pkg := range queryResponse {
		availablePkgs[pkg.Pname] = true
	}

	for _, pkg := range a.Packages {
		if _, present := availablePkgs[pkg]; !present {
			return false, nil
		}
	}

	return true, nil
}

// Push does nothing because this target does not push artifact to a remote cache
func (a Artifact) Push(_ context.Context) error {
	log.Debug("The nix target does not use a remote cache, so an artifact push will do nothing and return no error")
	return nil
}

// Pull does nothing because this target will never pass a remote cache check
func (a Artifact) Pull(_ context.Context) error {
	log.Debug("The nix target does not use a remote cache, so an artifact pull will do nothing and return no error")
	return nil
}
