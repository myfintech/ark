package stdlib

import (
	"github.com/dop251/goja"
	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/ark/targets/docker_image"
	"github.com/myfintech/ark/src/go/lib/ark/targets/group"
	"github.com/myfintech/ark/src/go/lib/ark/targets/kube_exec"
	"github.com/myfintech/ark/src/go/lib/ark/targets/local_file"
	"github.com/myfintech/ark/src/go/lib/ark/targets/nix"
	"github.com/myfintech/ark/src/go/lib/ark/targets/probe"
	"github.com/myfintech/ark/src/go/lib/ark/targets/sync_kv"
	"github.com/myfintech/ark/src/go/lib/ark/targets/test"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript/runtime/helpers"
)

// NewArkActionLibrary register all TS-Golang mappings
func NewArkActionLibrary(opts Options) typescript.Module {
	return typescript.Module{
		"buildDockerImage": NewAddTargetFunc(docker_image.Type, opts),
		"deploy":           NewAddTargetFunc("deploy", opts),
		"syncKV":           NewAddTargetFunc(sync_kv.Type, opts),
		"kubeExec":         NewAddTargetFunc(kube_exec.Type, opts),
		"group":            NewAddTargetFunc(group.Type, opts),
		"probe":            NewAddTargetFunc(probe.Type, opts),
		"test":             NewAddTargetFunc(test.Type, opts),
		"localFile":        NewAddTargetFunc(local_file.Type, opts),
		"nix":              NewAddTargetFunc(nix.Type, opts),
		"connectTargets":   NewConnectTargetFunc(opts),
	}
}

// TODO: all this function should be private

// NewConnectTargetFunc connects to a target function
func NewConnectTargetFunc(opts Options) func(call goja.FunctionCall) goja.Value {
	return helpers.NewGojaErrHandler(opts.Runtime, func(call goja.FunctionCall) (goja.Value, error) {
		v := call.Argument(0)
		edge := ark.GraphEdge{}

		err := opts.Runtime.ExportTo(v, &edge)
		if err != nil {
			return nil, err
		}

		_, err = opts.Client.ConnectTargets(edge)
		if err != nil {
			return nil, err
		}

		return opts.Runtime.ToValue(edge), nil
	})
}

// NewAddTargetFunc adds a new target
func NewAddTargetFunc(targetType string, opts Options) func(call goja.FunctionCall) goja.Value {
	return helpers.NewGojaErrHandler(opts.Runtime, func(call goja.FunctionCall) (goja.Value, error) {
		v := call.Argument(0)
		rawTarget := ark.RawTarget{
			Type:  targetType,
			Realm: opts.FSRealm,
			File:  helpers.GetRootCaller(opts.Runtime),
		}

		err := opts.Runtime.ExportTo(v, &rawTarget)
		if err != nil {
			return nil, err
		}

		artifact, err := opts.Client.AddTarget(rawTarget)
		if err != nil {
			return nil, err
		}

		return opts.Runtime.ToValue(artifact), nil
	})
}
