package cmd

import (
	"github.com/spf13/cobra"
)

func newServerCmd(rootCmd *cobra.Command) *cobra.Command {
	var serverCmd = &cobra.Command{
		Use:   "server",
		Short: "server is a sub-command of ark that starts the host server",
	}

	rootCmd.AddCommand(serverCmd)
	return serverCmd
}
