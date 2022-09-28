package sync_kv

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/myfintech/ark/src/go/lib/vault_tools"

	vault "github.com/hashicorp/vault/api"

	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/ark/kv"
	"github.com/stretchr/testify/require"

	"github.com/myfintech/ark/src/go/lib/vault_tools/vault_test_harness"
)

func TestKVSync(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	client, cleanup := vault_test_harness.CreateVaultTestCore(t, false)
	defer cleanup()

	testdata := filepath.Join(cwd, "testdata")

	target := &Target{
		Engine:         "vault",
		EngineURL:      client.Address(),
		TimeoutSeconds: 30,
		Token:          client.Token(),
		RawTarget: ark.RawTarget{
			Name:  "test",
			Type:  Type,
			File:  "test",
			Realm: testdata,
			SourceFiles: []string{
				filepath.Join(testdata, ".ark/kv/secret/foo"),
			},
		},
	}

	kvStore := &kv.VaultStorage{
		Client:        client,
		FSBasePath:    filepath.Join(testdata, ".ark/kv"),
		EncryptionKey: "domain-key",
	}

	// write some data to ark's KV store
	secretData, err := kvStore.Put("secret/foo", map[string]interface{}{
		"fruit": "apple",
	})
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(filepath.Join(testdata, ".ark/kv"))
	}()

	checksum, err := target.Checksum()
	require.NoError(t, err)

	artifact, err := target.Produce(checksum)
	require.NoError(t, err)

	action := &Action{
		Target:    target,
		KVStorage: kvStore,
		Artifact:  artifact.(*Artifact),
		VaultClientFactory: func(config *vault.Config) (*vault.Client, error) {
			return client, nil
		},
	}

	require.Implements(t, (*ark.Action)(nil), action)

	require.NoError(t, target.Validate())

	// sync data from the ark KV store to the Vault cluster
	err = action.Execute(context.Background())
	require.NoError(t, err)

	// result := artifact.(*Artifact)
	// require.False(t, result.Cacheable())

	// verify that the Vault cluster has the data that was synchronized
	secret, err := client.Logical().Read(vault_tools.SecretDataPath("secret/foo"))
	require.NoError(t, err)
	require.NotNil(t, secret)
	require.Equal(t, secretData, secret.Data["data"])
}
