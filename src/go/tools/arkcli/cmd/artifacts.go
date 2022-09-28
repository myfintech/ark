package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var artifactsCmd = &cobra.Command{
	Use:               "artifacts",
	Short:             "artifacts is a sub-command of ark that executes actions against buildable artifacts",
	PersistentPreRunE: decodeWorkspacePreRunE,
}

func init() {
	rootCmd.AddCommand(artifactsCmd)
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
}
