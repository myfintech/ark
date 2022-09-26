package kv

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/myfintech/ark/src/go/lib/vault_tools/vault_test_harness"

	"github.com/stretchr/testify/require"
)

func TestVaultStorage(t *testing.T) {
	path := "secret/foo"

	cwd, err := os.Getwd()
	require.NoError(t, err)

	client, cleanup := vault_test_harness.CreateVaultTestCore(t, false)
	defer cleanup()

	storage := VaultStorage{
		Client:        client,
		FSBasePath:    filepath.Join(cwd, "testdata", ".ark/kv"),
		EncryptionKey: "mantl-key",
	}

	defer func() {
		_ = os.RemoveAll(storage.FSBasePath)
	}()

	_, err = storage.Put(path, map[string]interface{}{
		"foo": "bar",
		"bar": "baz",
	})
	require.NoError(t, err)

	actualConfig, err := storage.Get(path)
	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{
		"foo": "bar",
		"bar": "baz",
	}, actualConfig)

	_, err = storage.Put(path, map[string]interface{}{
		"fruit":     "apple",
		"vegetable": "carrot",
	})
	require.NoError(t, err)

	updatedConfig, err := storage.Get(path)
	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{
		"foo":       "bar",
		"bar":       "baz",
		"fruit":     "apple",
		"vegetable": "carrot",
	}, updatedConfig)

	tmpFile, err := storage.DecryptToFile(path)
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(tmpFile)
	}()
	require.True(t, strings.HasPrefix(tmpFile, os.TempDir()))

	var decryptedData map[string]interface{}
	decryptedBytes, err := ioutil.ReadFile(tmpFile)
	require.NoError(t, err)

	err = json.Unmarshal(decryptedBytes, &decryptedData)
	require.NoError(t, err)
	require.Equal(t, decryptedData, updatedConfig)
}
