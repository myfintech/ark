package kube_exec

import (
	"os"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/exec"
	"github.com/myfintech/ark/src/go/lib/hclutils"
	"github.com/zclconf/go-cty/cty"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Target an executable target, when built it runs the specified command
type Target struct {
	*base.RawTarget `json:"-"`

	ResourceType  hcl.Expression `hcl:"resource_type,attr"`
	ResourceName  hcl.Expression `hcl:"resource_name,attr"`
	ContainerName hcl.Expression `hcl:"container_name,optional"`
	Command       hcl.Expression `hcl:"command,attr"`
	GetPodTimeout hcl.Expression `hcl:"get_pod_timeout,optional"`
}

// ComputedAttrs used to store the computed attributes of a kube_exec target
type ComputedAttrs struct {
	ResourceType  string   `hcl:"resource_type,attr"`
	ResourceName  string   `hcl:"resource_name,attr"`
	ContainerName string   `hcl:"container_name,optional"`
	Command       []string `hcl:"command,attr"`
	GetPodTimeout string   `hcl:"get_pod_timeout,optional"`
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

	var getPodTimeout time.Duration

	if attrs.GetPodTimeout == "" {
		getPodTimeout = 10 * time.Second
	} else {
		dur, err := time.ParseDuration(attrs.GetPodTimeout)
		if err != nil {
			return err
		}
		getPodTimeout = dur
	}

	pod, err := t.Workspace.K8s.GetPodByResource(
		attrs.ResourceType,
		attrs.ResourceName,
		getPodTimeout,
	)

	if err != nil {
		return err
	}

	return exec.KubernetesExecutor(exec.KubernetesExecOptions{
		Factory:       t.Workspace.K8s.Factory,
		Pod:           pod,
		Executor:      exec.DefaultKubernetesExecutor(),
		ContainerName: attrs.ContainerName,
		Command:       attrs.Command,
		IOStreams: genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
		GetPodTimeout: getPodTimeout,
	})
}
