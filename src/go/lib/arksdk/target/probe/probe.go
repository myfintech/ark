package probe

import (
	"net/url"
	"time"

	intlnet "github.com/myfintech/ark/src/go/lib/internal_net"
	"github.com/myfintech/ark/src/go/lib/log"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/hclutils"
)

// Target blocks the build walk until the specified address is reachable or fails
type Target struct {
	*base.RawTarget `json:"-"`

	DialAddress    hcl.Expression `hcl:"address,attr"`
	Timeout        hcl.Expression `hcl:"timeout,optional"`
	Delay          hcl.Expression `hcl:"delay,optional"`
	MaxRetries     hcl.Expression `hcl:"max_retries,optional"`
	ExpectedStatus hcl.Expression `hcl:"expected_status,optional"`
}

// ComputedAttrs used to store the computed attributes of a probe target
type ComputedAttrs struct {
	DialAddress    string `hcl:"address,attr"`
	Timeout        string `hcl:"timeout,optional"`
	Delay          string `hcl:"delay,optional"`
	MaxRetries     int    `hcl:"max_retries,optional"`
	ExpectedStatus int    `hcl:"expected_status,optional"`
}

// Attributes returns a combined map of rawTarget.Attributes and typedTarget.Attributes
func (t Target) Attributes() map[string]cty.Value {
	// computed := t.ComputedAttrs()
	return hclutils.MergeMapStringCtyValue(t.RawTarget.Attributes(), map[string]cty.Value{})
}

// ComputedAttrs returns a pointer to computed attributes from the state store.
// If attributes are not in the state store it will create a new pointer and insert it into the state store.
func (t Target) ComputedAttrs() *ComputedAttrs {
	if attrs, ok := t.GetStateAttrs().(*ComputedAttrs); ok {
		return attrs
	}

	attrs := &ComputedAttrs{}

	if attrs.Timeout == "" {
		attrs.Timeout = "5s"
	}

	if attrs.Delay == "" {
		attrs.Delay = "1s"
	}

	if attrs.MaxRetries == 0 {
		attrs.MaxRetries = 5
	}

	t.SetStateAttrs(attrs)
	return attrs
}

// PreBuild a lifecycle hook for calculating state before the build
func (t Target) PreBuild() error {
	attrs := t.ComputedAttrs()
	err := hclutils.DecodeExpressions(&t, attrs, base.CreateEvalContext(base.EvalContextOptions{
		CurrentTarget:     t,
		Package:           *t.Package,
		TargetLookupTable: t.Workspace.TargetLUT,
		Workspace:         *t.Workspace,
	}))
	if err != nil {
		return err
	}
	return nil
}

// Build creates the file from the content in the location specified with the permissions specified
func (t Target) Build() error {
	attrs := t.ComputedAttrs()

	timeout, err := time.ParseDuration(attrs.Timeout)
	if err != nil {
		return err
	}

	delay, err := time.ParseDuration(attrs.Delay)
	if err != nil {
		return err
	}

	addr, err := url.Parse(attrs.DialAddress)
	if err != nil {
		return err
	}

	probe, options, err := intlnet.CreateProbe(intlnet.ProbeOptions{
		Timeout:        timeout,
		Delay:          delay,
		Address:        addr,
		MaxRetries:     attrs.MaxRetries,
		ExpectedStatus: attrs.ExpectedStatus,
		OnError: func(err error, remainingAttempts int) {
			log.Warnf("an error occurred while running probe %s(%s), remaining attempts %d, %v",
				t.Address(), attrs.DialAddress, remainingAttempts, err)
		},
	})

	if err != nil {
		return err
	}

	return intlnet.RunProbe(options, probe)
}

// CacheEnabled overrides the default target caching behavior
func (t Target) CacheEnabled() {
	return
}
