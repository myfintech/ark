package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var kvCmd = &cobra.Command{
	Use:               "kv",
	Short:             "kv is a sub-command of ark that manages configurations for local Vault",
	PersistentPreRunE: decodeWorkspaceOnlyPreRunE,
}

func init() {
	rootCmd.AddCommand(kvCmd)
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
}
