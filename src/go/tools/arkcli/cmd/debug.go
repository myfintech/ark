package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var debugCmd = &cobra.Command{
	Use:               "debug",
	Short:             "debug is a sub-command of ark that outputs debug info on the specified object",
	PersistentPreRunE: decodeWorkspacePreRunE,
}

func init() {
	rootCmd.AddCommand(debugCmd)
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
}
