package docker_exec

import (
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/exec"
	"github.com/myfintech/ark/src/go/lib/hclutils"
)

// Target an executable target, when built it runs the specified command
type Target struct {
	*base.RawTarget `json:"-"`

	Command          hcl.Expression `hcl:"command,attr"`
	Image            hcl.Expression `hcl:"image,attr"`
	Environment      hcl.Expression `hcl:"environment,attr"`
	Volumes          hcl.Expression `hcl:"volumes,attr"`
	WorkingDirectory hcl.Expression `hcl:"working_directory,attr"`
	Ports            hcl.Expression `hcl:"ports,optional"`
	Detach           hcl.Expression `hcl:"detach,optional"`
	KillTimeout      hcl.Expression `hcl:"kill_timeout,optional"`
	Privileged       hcl.Expression `hcl:"privileged,optional"`
	NetworkMode      hcl.Expression `hcl:"network_mode,optional"`
}

// ComputedAttrs used to store the computed attributes of a local_exec target
type ComputedAttrs struct {
	Command          []string           `hcl:"command,attr"`
	Image            string             `hcl:"image,attr"`
	Volumes          *[]string          `hcl:"volumes,attr"`
	Environment      *map[string]string `hcl:"environment,attr"`
	WorkingDirectory *string            `hcl:"working_directory,attr"`
	Ports            []string           `hcl:"ports,optional"`
	Detach           bool               `hcl:"detach,optional"`
	KillTimeout      string             `hcl:"kill_timeout,optional"`
	Privileged       bool               `hcl:"privileged,optional"`
	NetworkMode      string             `hcl:"network_mode,optional"`
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

// WorkingDir the working directory of the container
func (t Target) WorkingDir() string {
	attrs := t.ComputedAttrs()
	if attrs.WorkingDirectory == nil {
		return t.Dir
	}
	return *attrs.WorkingDirectory
}

// Build executes the command specified in this target
func (t Target) Build() error {
	attrs := t.ComputedAttrs()
	return exec.DockerExecutor(nil, exec.DockerExecOptions{
		Command:          attrs.Command,
		Image:            attrs.Image,
		Dir:              t.WorkingDir(),
		Ports:            attrs.Ports,
		Detach:           attrs.Detach,
		ContainerName:    t.Address(),
		Binds:            *attrs.Volumes,
		Environment:      *attrs.Environment,
		Stdin:            os.Stdin,
		Stdout:           os.Stdout,
		Stderr:           os.Stderr,
		InheritParentEnv: true,
		KillTimeout:      attrs.KillTimeout,
		Privileged:       attrs.Privileged,
		NetworkMode:      attrs.NetworkMode,
	})
}
