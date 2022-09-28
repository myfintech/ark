package kv_sync

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	vault "github.com/hashicorp/vault/api"

	"github.com/myfintech/ark/src/go/lib/vault_tools"

	"github.com/myfintech/ark/src/go/lib/hclutils"

	"github.com/myfintech/ark/src/go/lib/ark/kv"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"

	"github.com/myfintech/ark/src/go/lib/vault_tools/vault_test_harness"

	"github.com/stretchr/testify/require"
)

const kvTargetRawHCL = `
	package "dev" {
		description = "kv_sync testing"
	}

	target "kv_sync" "seed_vault" {
        engine = "vault"
        engine_url = "%s"
        timeout = "30s"
        token = "%s"

        source_files = [
            "${workspace.path}/.ark/kv/secret/foo"
        ]
     }
`

func TestKVSync(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	client, cleanup := vault_test_harness.CreateVaultTestCore(t, false)
	defer cleanup()

	testDataDir := filepath.Join(cwd, "testdata")

	workspace := base.NewWorkspace()
	workspace.Dir = testDataDir

	workspace.RegisteredTargets = base.Targets{
		"kv_sync": Target{},
	}

	workspace.Vault = client

	workspace.VaultClientFactory = func(_ *vault.Config) (*vault.Client, error) {
		return client, nil
	}

	workspace.KVStorage = &kv.VaultStorage{
		Client:        workspace.Vault,
		FSBasePath:    filepath.Join(workspace.Dir, ".ark/kv"),
		EncryptionKey: "domain-key",
	}

	secretData, err := workspace.KVStorage.Put("secret/foo", map[string]interface{}{
		"fruit": "apple",
	})
	require.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(filepath.Join(workspace.Dir, ".ark/kv"))
	}()

	require.NoError(t, workspace.DetermineRoot(testDataDir))

	diag := workspace.DecodeFile(nil)
	require.False(t, diag.HasErrors(), diag.Error())

	hclFile, diag := hclutils.FileFromString(fmt.Sprintf(kvTargetRawHCL, client.Address(), client.Token()))
	require.False(t, diag.HasErrors(), diag.Error())

	err = workspace.LoadTargets([]base.BuildFile{{
		HCL:  hclFile,
		Path: "test",
	}})
	require.NoError(t, err)

	vertex, err := workspace.TargetLUT.LookupByAddress("dev.kv_sync.seed_vault")
	require.NoError(t, err)

	target := vertex.(Target)

	err = target.PreBuild()
	require.NoError(t, err)

	err = target.Build()
	require.NoError(t, err)

	secret, err := client.Logical().Read(vault_tools.SecretDataPath("secret/foo"))
	require.NoError(t, err)
	require.NotNil(t, secret)
	require.Equal(t, secretData, secret.Data["data"])
}
