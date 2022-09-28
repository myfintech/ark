package stdlib

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBase64FuncFunc(t *testing.T) {
	mod, err := vm.ResolveModule(filepath.Join(testdata, "src/encoding/encode.ts"))
	require.NoError(t, err)
	require.NotPanics(t, func() {
		v := mod.Get("exports").Export().(map[string]interface{})
		encoded := v["encoded"].(string)
		require.NotEmpty(t, encoded)
		require.Equal(t, encoded, "bXlWYWx1ZQ==")
	})

	mod, err = vm.ResolveModule(filepath.Join(testdata, "src/encoding/decode.ts"))
	require.NoError(t, err)
	require.NotPanics(t, func() {
		v := mod.Get("exports").Export().(map[string]interface{})
		decoded := v["decoded"].(string)
		require.NotEmpty(t, decoded)
		require.Equal(t, decoded, "myValue")
	})
}

func TestJson2stringFunc(t *testing.T) {
	mod, err := vm.ResolveModule(filepath.Join(testdata, "src/encoding/json.ts"))
	require.NoError(t, err)
	require.NotPanics(t, func() {
		v := mod.Get("exports").Export().(map[string]interface{})
		encoded := v["encoded"].(string)
		require.NotEmpty(t, encoded)
		require.Equal(t, encoded, "{\"key3\":\"k3\",\"key1\":\"k1\",\"key2\":\"k2\"}")
	})
}
