package plugins

import (
	"github.com/dop251/goja"
	"github.com/myfintech/ark/src/go/lib/ark/scripting/typescript/stdlib"
	"github.com/myfintech/ark/src/go/lib/arkplugins/kube-state/pkg"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript/runtime/helpers"
	"github.com/pkg/errors"
)

func NewKubeStatePlugin(opts stdlib.Options) NativePlugin {
	return NativePlugin{
		Name: "@mantl/sre/kube-state",
		Module: typescript.Module{
			"default": generateKubeStateManifest(opts),
		},
	}
}

func generateKubeStateManifest(opts stdlib.Options) func(call goja.FunctionCall) goja.Value {
	return helpers.NewGojaErrHandler(opts.Runtime, func(call goja.FunctionCall) (goja.Value, error) {
		pluginOpts := pkg.KubeStateOptions{
			Name:     "kubernetes-state",
			Replicas: 3,
		}

		if err := opts.Runtime.ExportTo(call.Argument(0), &pluginOpts); err != nil {
			return goja.Null(), errors.Wrap(err, "failed to export options from goja runtime")
		}

		manifest, err := pkg.NewKubeStateManifest(pluginOpts)
		if err != nil {
			return goja.Null(), errors.Wrap(err, "failed to generate vault manifest")
		}
		return opts.Runtime.ToValue(manifest), nil
	})
}
