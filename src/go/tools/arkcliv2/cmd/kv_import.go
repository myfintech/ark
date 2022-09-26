package cmd

import (
	"strings"

	"github.com/myfintech/ark/src/go/lib/ark/workspace"

	"github.com/myfintech/ark/src/go/lib/ark/kv"

	vault "github.com/hashicorp/vault/api"

	"github.com/spf13/cobra"

	"github.com/myfintech/ark/src/go/lib/log"
	"github.com/myfintech/ark/src/go/lib/vault_tools"
)

func newKVImportCmd(
	kvCmd *cobra.Command,
	vaultClient *vault.Client,
	storage kv.Storage,
	config *workspace.Config,
) *cobra.Command {
	var kvImportCmd = &cobra.Command{
		Use:     "import STARTING_VAULT_PATH",
		Short:   "import is a sub-command of kv that recursively traverses a given starting path in vault and uses the transit secrets engine to encrypt the content locally",
		PreRunE: requireKVSecretPath(config),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			log.Infof("crawling %s, this might take a moment", path)
			secretPaths, err := vault_tools.FindAllSecretsRecursive(vaultClient, []string{path})
			if err != nil {
				return err
			}

			for _, secretPath := range secretPaths {
				log.Infof("importing %s", secretPath)
				secret, sErr := vaultClient.Logical().Read(secretPath)
				if sErr != nil {
					return sErr
				}

				secretData, ok := secret.Data["data"].(map[string]interface{})
				if !ok {
					log.Warnf("skipping %s (may be marked as deleted)", secretPath)
					continue
				}

				_, err = storage.Put(strings.TrimPrefix(secretPath, "secret/data"), secretData)
				if err != nil {
					return err
				}
			}

			log.Info("done!")
			return nil
		},
	}

	kvCmd.AddCommand(kvImportCmd)
	kvImportCmd.Flags().StringP("file", "f", "", "The path to a JSON encoded file to import into the given path")
	return kvImportCmd
}
