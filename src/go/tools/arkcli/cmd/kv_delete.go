package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var kvDeleteCmd = &cobra.Command{
	Use:   "delete SECRET_PATH",
	Short: "delete is a sub-command of kv that deletes a kv file",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("SECRET_PATH is required")
		}

		if workspace.Config.Vault == nil || workspace.Config.Vault.EncryptionKey == "" {
			return errors.New("must specify vault.encryption_key in WORKSPACE.hcl")
		}
		return nil
	},
	ValidArgsFunction: getValidKVFiles,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		return os.Remove(filepath.Join(workspace.KVStorage.EncryptedDataPath(), path))
	},
}

func init() {
	kvCmd.AddCommand(kvDeleteCmd)
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
}
