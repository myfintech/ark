package typescript

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

func TestModuleResolver(t *testing.T) {
	vm := goja.New()
	transpiler, err := NewTranspiler()
	require.NoError(t, err)

	cwd, err := os.Getwd()
	require.NoError(t, err)

	testdata := filepath.Join(cwd, "testdata", "03_resolver")
	entrypoint := filepath.Join(testdata, "main.ts")

	resolver := DefaultModuleResolver(transpiler)
	resolver.Libraries = []Library{
		{
			Prefix: "ark/external", Path: filepath.Join(testdata, "lib", "ark_modules", "ark"),
		},
	}

	err = InstallPlugins(vm, []Plugin{
		transpiler,
		resolver,
	})
	require.NoError(t, err)

	mainModule, err := resolver.resolveAndTranspile(entrypoint)
	require.NoError(t, err)

	export := mainModule.Get("exports").Export().(map[string]interface{})
	require.Contains(t, export, "default")
	require.Contains(t, export, "example")
	require.Contains(t, export["default"], "example")
	require.Equal(t, export["example"], "test")
}
func TestModuleResolverShouldPreferAbsPath(t *testing.T) {
	vm := goja.New()
	transpiler, err := NewTranspiler()
	require.NoError(t, err)

	cwd, err := os.Getwd()
	require.NoError(t, err)

	testdata := filepath.Join(cwd, "testdata", "04_prefer_abs_file_over_folder")
	entrypoint := filepath.Join(testdata, "index.ts")

	resolver := DefaultModuleResolver(transpiler)

	err = InstallPlugins(vm, []Plugin{
		transpiler,
		resolver,
	})
	require.NoError(t, err)

	mainModule, err := resolver.resolveAndTranspile(entrypoint)
	require.NoError(t, err)

	results := make(map[string]string)

	module := mainModule.Get("exports").ToObject(vm).Get("default")
	err = vm.ExportTo(module, &results)
	require.NoError(t, err)

	require.Equal(t, "mod.ts", results["mod"])
	require.Equal(t, "mod/index.ts", results["idx"])
}

func TestEntrypointResolution(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	testdata := filepath.Join(cwd, "testdata")
	modpath := filepath.Join(testdata, "04_prefer_abs_file_over_folder", "mod.config")
	expected := filepath.Join(testdata, "04_prefer_abs_file_over_folder", "mod.config.ts")

	found, err := resolveModuleFile(modpath)
	require.NoError(t, err)
	require.Equal(t, expected, found)
}
