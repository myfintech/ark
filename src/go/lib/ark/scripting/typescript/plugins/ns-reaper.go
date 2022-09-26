package plugins

import (
	"github.com/dop251/goja"
	"github.com/myfintech/ark/src/go/lib/ark/scripting/typescript/stdlib"
	"github.com/myfintech/ark/src/go/lib/arkplugins/ns-reaper/pkg"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript/runtime/helpers"
	"github.com/myfintech/ark/src/go/lib/kube/microservice"
	"github.com/pkg/errors"
)

func NewNsReaperPlugin(opts stdlib.Options) NativePlugin {
	return NativePlugin{
		Name: "@mantl/sre/ns-reaper",
		Module: typescript.Module{
			"default": generateNsReaperManifest(opts),
		},
	}
}

func generateNsReaperManifest(opts stdlib.Options) func(call goja.FunctionCall) goja.Value {
	return helpers.NewGojaErrHandler(opts.Runtime, func(call goja.FunctionCall) (goja.Value, error) {
		pluginOpts := microservice.Options{}

		if err := opts.Runtime.ExportTo(call.Argument(0), &pluginOpts); err != nil {
			return goja.Null(), errors.Wrap(err, "failed to export options from goja runtime")
		}

		manifest, err := pkg.NewNsReaperManifest(pluginOpts)
		if err != nil {
			return goja.Null(), errors.Wrap(err, "failed to generate vault manifest")
		}
		return opts.Runtime.ToValue(manifest), nil
	})
}
