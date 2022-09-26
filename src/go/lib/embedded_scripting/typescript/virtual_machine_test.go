package typescript

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

type vmPlugin struct{}

func (vmPlugin) Install(runtime *goja.Runtime) error {
	return runtime.Set("echo", func(call goja.FunctionCall) goja.Value {
		return call.Argument(0)
	})
}

func TestNewVirtualMachine(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	testdata := filepath.Join(cwd, "testdata", "03_resolver")

	vm, err := NewVirtualMachine([]Library{
		{
			Prefix: "ark/external",
			Path:   filepath.Join(testdata, "lib", "ark_modules", "ark"),
		},
	})
	require.NoError(t, err)

	entrypoint := filepath.Join(testdata, "main.ts")

	mod, err := vm.ResolveModule(entrypoint)
	require.NoError(t, err)
	require.Contains(t, mod.Get("exports").Export(), "default")

	err = vm.InstallPlugins([]Plugin{
		vmPlugin{},
	})
	require.NoError(t, err)

	v, err := vm.RunScript("", "echo(10)")
	require.NoError(t, err)
	require.Equal(t, int64(10), v.ToInteger())
}
