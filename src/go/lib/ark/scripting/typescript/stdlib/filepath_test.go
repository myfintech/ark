package stdlib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGlobFunc(t *testing.T) {
	mod, err := vm.ResolveModule(filepath.Join(testdata, "src/fs/fs.ts"))
	require.NoError(t, err)
	require.NotPanics(t, func() {
		v := mod.Get("exports").Export().(map[string]interface{})
		files := v["files"].([]string)
		require.NotEmpty(t, files)
		for _, file := range files {
			info, err := os.Stat(file)
			require.NoError(t, err)
			require.Truef(t, !info.IsDir(), "%s should not be a directory", file)
		}
	})
}
