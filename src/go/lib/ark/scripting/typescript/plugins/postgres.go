package plugins

import (
	"github.com/dop251/goja"
	"github.com/myfintech/ark/src/go/lib/ark/scripting/typescript/stdlib"
	"github.com/myfintech/ark/src/go/lib/arkplugins/postgres/pkg"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript/runtime/helpers"
	"github.com/myfintech/ark/src/go/lib/kube/statefulapp"
	"github.com/pkg/errors"
)

func NewPostgresPlugin(opts stdlib.Options) NativePlugin {
	return NativePlugin{
		Name: "@mantl/sre/postgres",
		Module: typescript.Module{
			"default": generatePostgresManifest(opts),
		},
	}
}

func generatePostgresManifest(opts stdlib.Options) func(call goja.FunctionCall) goja.Value {
	return helpers.NewGojaErrHandler(opts.Runtime, func(call goja.FunctionCall) (goja.Value, error) {
		pluginOpts := pkg.Options{
			Options: statefulapp.Options{
				Replicas:    1,
				Port:        5432,
				ServicePort: 5432,
				Image:       "postgres:9",
				ServiceType: "ClusterIP",
				Env: map[string]string{
					"POSTGRES_PASSWORD": "mantl",
				},
			},
		}

		if err := opts.Runtime.ExportTo(call.Argument(0), &pluginOpts); err != nil {
			return goja.Null(), errors.Wrap(err, "failed to export options from goja runtime")
		}

		manifest, err := pkg.NewPostgresManifest(pluginOpts)
		if err != nil {
			return goja.Null(), errors.Wrap(err, "failed to generate vault manifest")
		}
		return opts.Runtime.ToValue(manifest), nil
	})
}
