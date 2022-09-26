package local_exec

import (
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/myfintech/ark/src/go/lib/log"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/exec"
	"github.com/myfintech/ark/src/go/lib/hclutils"
)

// Target an executable target, when built it runs the specified command
type Target struct {
	*base.RawTarget `json:"-"`

	Command     hcl.Expression `hcl:"command,attr"`
	Environment hcl.Expression `hcl:"environment,attr"`
}

// ComputedAttrs used to store the computed attributes of a local_exec target
type ComputedAttrs struct {
	Command     []string           `hcl:"command,attr"`
	Environment *map[string]string `hcl:"environment,attr"`
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
	cmd := exec.LocalExecutor(exec.LocalExecOptions{
		Command:          attrs.Command,
		Dir:              t.Dir,
		Environment:      *attrs.Environment,
		Stdin:            os.Stdin,
		Stdout:           os.Stdout,
		Stderr:           os.Stderr,
		InheritParentEnv: true,
	})
	log.Info(cmd.String())
	return cmd.Run()
}
