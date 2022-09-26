package plugins

import (
	"github.com/dop251/goja"
	"github.com/myfintech/ark/src/go/lib/ark/scripting/typescript/stdlib"
	"github.com/myfintech/ark/src/go/lib/arkplugins/vanity-domain/pkg"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript/runtime/helpers"
	"github.com/myfintech/ark/src/go/lib/kube/mutations"
	"github.com/pkg/errors"
)

// NewVDSPlugin creates a new instance of the VDS deployment plugin
func NewVDSPlugin(opts stdlib.Options) NativePlugin {
	return NativePlugin{
		Name: "@mantl/sre/vanity-domain",
		Module: typescript.Module{
			"default": generateVDSManifest(opts),
		},
	}
}

func generateVDSManifest(opts stdlib.Options) func(call goja.FunctionCall) goja.Value {
	return helpers.NewGojaErrHandler(opts.Runtime, func(call goja.FunctionCall) (goja.Value, error) {
		pluginOpts := pkg.Options{
			VaultConfig: mutations.VaultConfig{
				Team:          "sre",
				App:           "vanity-domain",
				DefaultConfig: "vanity-domain/config",
				Address:       "http://vault.es.mantl.internal",
			},
			Replicas: int32(3),
		}

		if err := opts.Runtime.ExportTo(call.Argument(0), &pluginOpts); err != nil {
			return goja.Null(), errors.Wrap(err, "failed to export options from goja runtime")
		}

		manifest, err := pkg.NewManifest(pluginOpts)
		if err != nil {
			return goja.Null(), errors.Wrap(err, "failed to generate vault manifest")
		}
		return opts.Runtime.ToValue(manifest), nil
	})
}
