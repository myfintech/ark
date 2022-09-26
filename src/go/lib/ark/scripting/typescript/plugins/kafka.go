package plugins

import (
	"github.com/dop251/goja"
	"github.com/myfintech/ark/src/go/lib/ark/scripting/typescript/stdlib"
	"github.com/myfintech/ark/src/go/lib/arkplugins/kafka/pkg"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript/runtime/helpers"
	"github.com/myfintech/ark/src/go/lib/kube/statefulapp"
	"github.com/pkg/errors"
)

func NewKafkaPlugin(opts stdlib.Options) NativePlugin {
	return NativePlugin{
		Name: "@mantl/sre/kafka",
		Module: typescript.Module{
			"default": generateKafkaManifest(opts),
		},
	}
}

func generateKafkaManifest(opts stdlib.Options) func(call goja.FunctionCall) goja.Value {
	return helpers.NewGojaErrHandler(opts.Runtime, func(call goja.FunctionCall) (goja.Value, error) {

		pluginOpts := pkg.Options{
			Options: statefulapp.Options{},
		}

		if err := opts.Runtime.ExportTo(call.Argument(0), &pluginOpts); err != nil {
			return goja.Null(), errors.Wrap(err, "failed to export options from goja runtime")
		}

		manifest, err := pkg.NewKafkaManifest(pluginOpts)
		if err != nil {
			return goja.Null(), errors.Wrap(err, "failed to generate kafka manifest")
		}
		return opts.Runtime.ToValue(manifest), nil
	})
}
