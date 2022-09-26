package plugins

import (
	"bytes"

	"github.com/dop251/goja"
	"github.com/myfintech/ark/src/go/lib/ark/scripting/typescript/stdlib"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript/runtime/helpers"
	"github.com/myfintech/ark/src/go/lib/kube/microservice"
	"github.com/pkg/errors"
)

func NewMicroservicePlugin(opts stdlib.Options) NativePlugin {
	return NativePlugin{
		Name: "@mantl/sre/microservice",
		Module: typescript.Module{
			"default": generateMicroserviceManifest(opts),
		},
	}
}

func generateMicroserviceManifest(opts stdlib.Options) func(call goja.FunctionCall) goja.Value {
	return helpers.NewGojaErrHandler(opts.Runtime, func(call goja.FunctionCall) (goja.Value, error) {
		buff := new(bytes.Buffer)
		pluginOpts := microservice.Options{
			Replicas: 1,
		}

		if err := opts.Runtime.ExportTo(call.Argument(0), &pluginOpts); err != nil {
			return goja.Null(), errors.Wrap(err, "failed to export options from goja runtime")
		}

		manifest := microservice.NewMicroService(pluginOpts)

		if err := manifest.Serialize(buff); err != nil {
			return goja.Null(), errors.Wrap(err, "failed to serialize manifest")
		}

		return opts.Runtime.ToValue(buff.String()), nil
	})
}
