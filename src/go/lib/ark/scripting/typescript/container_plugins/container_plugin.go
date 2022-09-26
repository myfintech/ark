package container_plugins

import (
	"bytes"
	"context"
	"os"
	"path/filepath"

	"github.com/dop251/goja"
	"github.com/myfintech/ark/src/go/lib/ark/workspace"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript/runtime/helpers"
	"github.com/myfintech/ark/src/go/lib/exec"
	"github.com/pkg/errors"
)

// Load a collection of plugin into a TypeScript VM
func Load(ctx context.Context, vm *typescript.VirtualMachine, plugins []workspace.Plugin) error {
	for _, plugin := range plugins {
		if err := vm.InstallModule(filepath.Join("ark/plugins", plugin.Name), typescript.Module{
			"default": newDockerRunFunc(ctx, vm.Runtime, plugin),
		}); err != nil {
			return errors.Errorf("got an error installing modules, %v", err)
		}
	}

	return nil
}

func newDockerRunFunc(ctx context.Context, runtime *goja.Runtime, plugin workspace.Plugin) func(call goja.FunctionCall) goja.Value {
	return helpers.NewGojaErrHandler(runtime, func(call goja.FunctionCall) (goja.Value, error) {
		v := call.Argument(0)

		buf, err := v.ToObject(runtime).MarshalJSON()
		if err != nil {
			return nil, err
		}

		input := bytes.NewReader(buf)
		output := bytes.NewBuffer(nil)
		if err = exec.DockerExecutor(ctx, exec.DockerExecOptions{ // gets its own docker client.
			Image:       plugin.Image,
			Stdin:       input,
			Stdout:      output,
			Stderr:      os.Stderr,
			AttachStdIn: true,
		}); err != nil {
			return nil, errors.Wrapf(err, "failed to execute ark plugin: %s", plugin.Name)
		}
		if output.String() == "" {
			return nil, errors.New("ark plugin returned no value")
		}
		return runtime.ToValue(output.String()), nil
	})
}
