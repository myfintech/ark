package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var kvEditCmd = &cobra.Command{
	Use:   "edit SECRET_PATH",
	Short: "edit is a sub-command of kv that reads configuration into a temp file and opens the temp file in the user's $EDITOR for modification",
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

		return workspace.KVStorage.Edit(path)
	},
}

func init() {
	kvCmd.AddCommand(kvEditCmd)
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
}
