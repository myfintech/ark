package nix

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/hclutils"
)

// Target an executable target, when built it runs the specified command
type Target struct {
	*base.RawTarget `json:"-"`
	Packages        hcl.Expression `hcl:"packages,attr"`
}

// ComputedAttrs used to store the computed attributes of a nix target
type ComputedAttrs struct {
	Packages []string `hcl:"packages,attr"`
}

// NxPkg is a container for unmarshalled nix query responses
type NxPkg struct {
	Name    string `json:"name"`
	Pname   string `json:"pname"`
	Version string `json:"version"`
	System  string `json:"system"`
}

// Attributes return a combined map of rawTarget.Attributes and typedTarget.Attributes
func (t Target) Attributes() map[string]cty.Value {
	return hclutils.MergeMapStringCtyValue(t.RawTarget.Attributes(), map[string]cty.Value{})
}

// ComputedAttrs returns a pointer to computed attributes from the state store.
// If attributes are not in the state store it will create a new pointer and insert it into the state store.
func (t Target) ComputedAttrs() *ComputedAttrs {
	if attrs, ok := t.GetStateAttrs().(*ComputedAttrs); ok {
		return attrs
	}

	attrs := &ComputedAttrs{}
	t.SetStateAttrs(attrs)
	return attrs
}

// CacheEnabled overrides the default target caching behavior
func (t Target) CacheEnabled() {
	return
}

// PreBuild a lifecycle hook for calculating state before the build
func (t Target) PreBuild() error {
	return hclutils.DecodeExpressions(&t, t.ComputedAttrs(), base.CreateEvalContext(base.EvalContextOptions{
		CurrentTarget:     t,
		Package:           *t.Package,
		TargetLookupTable: t.Workspace.TargetLUT,
		Workspace:         *t.Workspace,
	}))
}

// Build executes the command specified in this target
func (t Target) Build() error {
	attrs := t.ComputedAttrs()

	for _, p := range attrs.Packages {
		cmd := exec.Command("nix-env", "-iA", p)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

// CheckLocalBuildCache checks the nix packages that are requested against those that are already installed
func (t Target) CheckLocalBuildCache() (bool, error) {
	attrs := t.ComputedAttrs()
	queryResponse := map[string]NxPkg{}

	out, err := exec.Command("nix-env", "--query", "--installed", "--json").Output()
	if err != nil {
		return false, err
	}

	if err = json.Unmarshal(out, &queryResponse); err != nil {
		return false, err
	}

	availablePkgs := make(map[string]bool)

	for _, i := range queryResponse {
		availablePkgs[i.Pname] = true
	}

	for _, p := range attrs.Packages {
		pkg := strings.Split(p, ".")
		if _, present := availablePkgs[pkg[1]]; !present {
			return false, nil
		}
	}

	return true, nil
}
