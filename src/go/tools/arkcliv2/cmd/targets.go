package cmd

import (
	"github.com/spf13/cobra"
)

func newTargetsCmd(rootCmd *cobra.Command) *cobra.Command {
	var targetsCmd = &cobra.Command{
		Use:   "targets",
		Short: "targets is a sub-command of ark that interacts with buildable targets",
	}

	rootCmd.AddCommand(targetsCmd)
	return targetsCmd
}
