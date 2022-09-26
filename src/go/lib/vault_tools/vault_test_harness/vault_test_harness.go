package vault_test_harness

import (
	"context"
	"testing"

	logicalKV "github.com/hashicorp/vault-plugin-secrets-kv"
	vaultApi "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/audit"
	"github.com/hashicorp/vault/builtin/audit/file"
	"github.com/hashicorp/vault/builtin/logical/transit"
	"github.com/hashicorp/vault/helper/namespace"
	vaulthttp "github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/sdk/helper/logging"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/hashicorp/vault/vault"
	"github.com/stretchr/testify/require"
)

func CreateVaultTestCore(t *testing.T, useV2KVStore bool) (*vaultApi.Client, func()) {
	coreConfig := &vault.CoreConfig{
		// EnableUI: true, Vault UI is not included in the test harness
		LogicalBackends: map[string]logical.Factory{
			"transit": transit.Factory,
		},
		AuditBackends: map[string]audit.Factory{
			"file": file.Factory,
		},
	}

	if useV2KVStore {
		coreConfig.LogicalBackends["kv"] = logicalKV.VersionedKVFactory
	}

	cluster := vault.NewTestCluster(t, coreConfig, &vault.TestClusterOptions{
		HandlerFunc: vaulthttp.Handler,
		Logger:      logging.NewVaultLogger(6),
	})
	cluster.Start()

	core := cluster.Cores[0].Core
	vault.TestWaitActive(t, core)
	client := cluster.Cores[0].Client

	createRequest := &vaultApi.TokenCreateRequest{
		TTL:       "1h",
		Renewable: new(bool),
	}
	*createRequest.Renewable = true
	secret, err := client.Auth().Token().Create(createRequest)
	require.NoError(t, err)
	client.SetToken(secret.Auth.ClientToken)

	if err = client.Sys().EnableAuditWithOptions("file", &vaultApi.EnableAuditOptions{
		Type: "file",
		Options: map[string]string{
			"file_path": "/dev/null",
		},
	}); err != nil {
		require.NoError(t, err)
	}

	if useV2KVStore {
		err = client.Sys().Unmount("secret/")
		require.NoError(t, err)

		if err = client.Sys().Mount("kv", &vaultApi.MountInput{
			Type: "kv-v2",
		}); err != nil {
			require.NoError(t, err)
		}
		kvReq := &logical.Request{
			Operation:   logical.UpdateOperation,
			ClientToken: cluster.RootToken,
			Path:        "sys/mounts/secret",
			Data: map[string]interface{}{
				"type":        "kv",
				"path":        "secret/",
				"description": "key/value secret storage",
				"options": map[string]string{
					"version": "2",
				},
			},
		}
		resp, err := core.HandleRequest(namespace.RootContext(context.TODO()), kvReq)
		if err != nil {
			t.Fatal(err)
		}
		if resp.IsError() {
			t.Fatal(err)
		}
	}

	if err = client.Sys().Mount("transit", &vaultApi.MountInput{
		Type: "transit",
	}); err != nil {
		require.NoError(t, err)
	}

	_, err = client.Logical().Write("transit/keys/mantl-key", map[string]interface{}{})
	require.NoError(t, err)

	return client, cluster.Cleanup
}
