package kv_sync

import (
	"net/url"
	"strings"
	"time"

	intlnet "github.com/myfintech/ark/src/go/lib/internal_net"
	"github.com/myfintech/ark/src/go/lib/log"
	"github.com/myfintech/ark/src/go/lib/vault_tools"

	"github.com/myfintech/ark/src/go/lib/fs"

	"github.com/pkg/errors"

	"github.com/hashicorp/hcl/v2"
	vault "github.com/hashicorp/vault/api"
	"github.com/zclconf/go-cty/cty"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/hclutils"
)

// Target an executable target, when built it runs the specified command
type Target struct {
	*base.RawTarget `json:"-"`
	Engine          hcl.Expression `hcl:"engine,attr"`
	EngineURL       hcl.Expression `hcl:"engine_url,attr"`
	Timeout         hcl.Expression `hcl:"timeout,attr"`
	Token           hcl.Expression `hcl:"token,attr"`
}

// ComputedAttrs used to store the computed attributes of a kv_sync target
type ComputedAttrs struct {
	Engine    string `hcl:"string,attr"`
	EngineURL string `hcl:"engine_url,attr"`
	Timeout   string `hcl:"timeout,attr"`
	Token     string `hcl:"token,attr"`
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

	timeout, err := time.ParseDuration(attrs.Timeout)
	if err != nil {
		return err
	}

	var clientFactory base.VaultClientFactory

	if t.Workspace.VaultClientFactory != nil {
		clientFactory = t.Workspace.VaultClientFactory
	} else {
		clientFactory = vault.NewClient
	}

	client, err := clientFactory(&vault.Config{
		Address: attrs.EngineURL,
		Timeout: timeout,
	})
	if err != nil {
		return err
	}

	client.SetToken(attrs.Token)

	addr, err := url.Parse(client.Address())
	if err != nil {
		return err
	}

	probeOptions := intlnet.ProbeOptions{
		Timeout:    timeout,
		Address:    addr,
		MaxRetries: 6,
		Delay:      time.Second * 3,
		OnError: func(err error, remainingAttempts int) {
			log.Warnf("waiting for vault to initialize and unseal: %v, remaining attempts %d",
				err, remainingAttempts)
		},
	}

	// https://www.vaultproject.io/api/system/health
	// 200 == initialized, unsealed, and active (normally)
	if err = intlnet.RunProbe(probeOptions, func() error {
		health, hErr := client.Sys().Health()
		if hErr != nil {
			return hErr
		}
		if health.Initialized && health.Sealed == false {
			return nil
		}
		return errors.New("not healthy")
	}); err != nil {
		return err
	}

	secretPrefix := t.Workspace.KVStorage.EncryptedDataPath()

	if len(t.SourceFilesList()) == 0 {
		return errors.New("must supply at least one secret source file")
	}

	matchCache, _ := t.FileDeps()
	for _, file := range matchCache.FilesList() {
		if file.Type != "f" {
			log.Debugf("skipping non-file %s %s", file.Type, file.Name)
			continue
		}

		if !strings.HasPrefix(file.Name, secretPrefix) {
			return errors.Errorf("the source file %s does not match the prefix: %s", file.Name, secretPrefix)
		}

		trimmedPath := fs.TrimPrefix(file.Name, secretPrefix)

		data, getErr := t.Workspace.KVStorage.Get(trimmedPath)
		if getErr != nil {
			return getErr
		}

		_, writeErr := client.Logical().Write(vault_tools.SecretDataPath(trimmedPath), map[string]interface{}{
			"data": data,
		})
		if writeErr != nil {
			return writeErr
		}
	}

	return nil
}
