package xdgbase

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/myfintech/ark/src/go/lib/fs"
)

func TestDir(t *testing.T) {
	cwd, _ := os.Getwd()
	testdata := filepath.Join(cwd, "testdata")
	// require.NoError(t, os.MkdirAll(testdata, 0777))
	// defer func() { _ = os.RemoveAll(testdata) }()

	testCases := []struct {
		vendor string
		suffix Suffix
	}{
		{
			vendor: "ark",
			suffix: CacheSuffix,
		},
		{
			vendor: "ark",
			suffix: DataSuffix,
		},
		{
			vendor: "ark",
			suffix: ConfigSuffix,
		},
	}

	for _, testCase := range testCases {
		vendor := testCase.vendor
		suffix := testCase.suffix
		t.Run(fmt.Sprintf("vendor %s should resolve by suffix %s", vendor, suffix), func(t *testing.T) {
			t.Run("should resolve to default $HOME directory", func(t *testing.T) {
				home, err := fs.NormalizePath(cwd, defaultDir(vendor, suffix))
				require.NoError(t, err)

				dir, err := Dir(vendor, suffix)
				require.NoError(t, err)
				require.Equal(t, home, dir)
			})

			t.Run("should resolve to XDG suffix", func(t *testing.T) {
				require.NoError(t, os.Setenv(vendorKey("xdg", suffix), filepath.Join(testdata, "xdg")))
				xdgDir, err := Dir(vendor, suffix)
				require.NoError(t, err)
				require.Equal(t, filepath.Join(testdata, "xdg"), xdgDir)
			})

			t.Run("should resolve to VENDOR suffix", func(t *testing.T) {
				require.NoError(t, os.Setenv(vendorKey(vendor, suffix), testdata))
				vDir, err := Dir(vendor, suffix)
				require.NoError(t, err)
				require.Equal(t, testdata, vDir)
				require.NoError(t, os.Unsetenv(vendorKey(vendor, suffix)))
			})

		})
	}

}
