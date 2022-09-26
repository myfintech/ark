package fs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArchive(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	testdata := filepath.Join(cwd, "testdata")
	testArchive := filepath.Join(testdata, "test.tar.gz")

	defer func() {
		_ = os.Remove(testArchive)
	}()

	dest, err := os.OpenFile(testArchive, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	require.NoError(t, err)
	defer func() {
		_ = dest.Close()
	}()

	require.NoError(t, GzipTar(testdata, dest))

	src, err := os.Open(testArchive)
	require.NoError(t, err)
	defer func() {
		_ = src.Close()
	}()

	require.NoError(t, GzipUntar(testdata, src))
}

func TestGzipTarFiles(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	testdata := filepath.Join(cwd, "testdata")
	testArchive := filepath.Join(testdata, "test.tar.gz")
	keepFile := filepath.Join(testdata, "keep_file_1.md")

	defer func() {
		_ = os.Remove(testArchive)
		_ = os.Remove(filepath.Join(testdata, "test.json"))
	}()

	dest, err := os.OpenFile(testArchive, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	require.NoError(t, err)
	defer func() {
		_ = dest.Close()
	}()

	jsonBytes, err := json.Marshal(map[string]interface{}{
		"test": true,
	})
	require.NoError(t, err)

	err = GzipTarFiles([]string{keepFile}, testdata, dest, InjectTarFiles([]*TarFile{
		{
			Name: "test.json",
			Body: jsonBytes,
			Mode: 0600,
		},
	}))
	require.NoError(t, err)

	src, err := os.Open(testArchive)
	require.NoError(t, err)
	defer func() {
		_ = src.Close()
	}()

	err = GzipUntar(testdata, src)
	require.NoError(t, err)
}
