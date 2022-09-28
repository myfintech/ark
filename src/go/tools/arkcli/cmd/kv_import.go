package cmd

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/myfintech/ark/src/go/lib/log"
	"github.com/myfintech/ark/src/go/lib/vault_tools"
)

var kvImportCmd = &cobra.Command{
	Use:   "import STARTING_VAULT_PATH",
	Short: "import is a sub-command of kv that recursively traverses a given starting path in vault and uses the transit secrets engine to encrypt the content locally",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("SECRET_PATH is a required parameter")
		}

		if workspace.Config.Vault == nil || workspace.Config.Vault.EncryptionKey == "" {
			return errors.New("must specify vault.encryption_key in WORKSPACE.hcl")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		log.Infof("crawling %s, this might take a moment", path)
		secretPaths, err := vault_tools.FindAllSecretsRecursive(workspace.Vault, []string{path})
		if err != nil {
			return err
		}

		for _, secretPath := range secretPaths {
			log.Infof("importing %s", secretPath)
			secret, sErr := workspace.Vault.Logical().Read(secretPath)
			if sErr != nil {
				return sErr
			}

			_, err = workspace.KVStorage.Put(strings.TrimPrefix(secretPath, "secret/data"), secret.Data["data"].(map[string]interface{}))
			if err != nil {
				return err
			}
		}

		log.Info("done!")
		return nil
	},
}

func init() {
	kvCmd.AddCommand(kvImportCmd)
	kvImportCmd.Flags().StringP("file", "f", "", "The path to a JSON encoded file to import into the given path")
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
}
