package vault_tools

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/myfintech/ark/src/go/lib/utils"
)

func TestFindAllSecretsRecursive(t *testing.T) {
	vaultClient, err := InitClient(nil)
	require.NoError(t, err)

	secretPaths, err := FindAllSecretsRecursive(vaultClient, []string{"oao", "integration", "trex"})
	require.NoError(t, err)

	t.Log(utils.MarshalJSONSafe(secretPaths, true))
	_, err = FindAllSecretsRecursive(vaultClient, []string{"non-exist"})
	require.Error(t, err, "should return an error when accessing an invalid path")
}
