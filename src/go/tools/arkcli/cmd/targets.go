package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var targetsCmd = &cobra.Command{
	Use:               "targets",
	Short:             "targets is a sub-command of ark that interacts with are buildable targets",
	PersistentPreRunE: decodeWorkspacePreRunE,
}

func init() {
	rootCmd.AddCommand(targetsCmd)
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
}
