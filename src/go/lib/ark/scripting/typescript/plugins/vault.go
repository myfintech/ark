package plugins

import (
	"github.com/dop251/goja"
	"github.com/myfintech/ark/src/go/lib/ark/scripting/typescript/stdlib"
	"github.com/myfintech/ark/src/go/lib/arkplugins/vault/pkg"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript/runtime/helpers"
	"github.com/myfintech/ark/src/go/lib/kube/statefulapp"
	"github.com/pkg/errors"
)

func NewVaultPlugin(opts stdlib.Options) NativePlugin {
	return NativePlugin{
		Name: "@mantl/sre/vault",
		Module: typescript.Module{
			"default": generateVaultManifest(opts),
		},
	}
}

func generateVaultManifest(opts stdlib.Options) func(call goja.FunctionCall) goja.Value {
	return helpers.NewGojaErrHandler(opts.Runtime, func(call goja.FunctionCall) (goja.Value, error) {
		pluginOpts := statefulapp.Options{
			Replicas:    1,
			Port:        8200,
			ServicePort: 8200,
			Image:       "gcr.io/managed-infrastructure/ark/dev/vault:f25bcc6e17016cc90cbf894be54dccaafb46231da65d1dbfdc2c855c30a949d6",
			ServiceType: "ClusterIP",
			Env:         make(map[string]string),
		}

		if err := opts.Runtime.ExportTo(call.Argument(0), &pluginOpts); err != nil {
			return goja.Null(), errors.Wrap(err, "failed to export options from goja runtime")
		}

		manifest, err := pkg.NewVaultManifest(pluginOpts)
		if err != nil {
			return goja.Null(), errors.Wrap(err, "failed to generate vault manifest")
		}
		return opts.Runtime.ToValue(manifest), nil
	})
}
