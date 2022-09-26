package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var validateCmd = &cobra.Command{
	Use:               "validate",
	Short:             "validate is a sub-command of ark that validates the WORKSPACE.hcl file",
	PersistentPreRunE: decodeWorkspacePreRunE,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
}
