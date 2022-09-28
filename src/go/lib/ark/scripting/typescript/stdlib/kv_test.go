package stdlib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetFunc(t *testing.T) {
	path := "secret/foo"
	store := filepath.Join(testdata, "src/kv/.ark/kv")

	defer func() {
		_ = os.RemoveAll(store)
	}()

	_, err := storage.Put(path, map[string]interface{}{
		"foo": "bar",
		"bar": "baz",
	})
	require.NoError(t, err)

	actualConfig, err := storage.Get(path)
	require.NoError(t, err)

	mod, err := vm.ResolveModule(filepath.Join(testdata, "src/kv/secrets.ts"))
	require.NoError(t, err)
	require.NotPanics(t, func() {
		v := mod.Get("exports").Export().(map[string]interface{})
		secret := v["secret"].(map[string]interface{})
		require.NotEmpty(t, secret)
		require.Equal(t, secret, actualConfig)
	})

}
