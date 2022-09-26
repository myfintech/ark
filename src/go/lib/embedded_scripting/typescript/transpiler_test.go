package typescript

import (
	"strings"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

func TestLoadCompiler(t *testing.T) {
	_, err := LoadCompiler()
	require.NoError(t, err)
}

func TestNewTranspiler(t *testing.T) {
	vm := goja.New()
	transpiler, err := NewTranspiler()
	require.NoError(t, err)

	err = InstallPlugins(vm, []Plugin{transpiler})
	require.NoError(t, err)

	source, err := transpiler.Transpile(strings.NewReader(`
	/* comment */
	export const example = { test: 123 }
	`))
	require.NoError(t, err)
	t.Log(source)
}
