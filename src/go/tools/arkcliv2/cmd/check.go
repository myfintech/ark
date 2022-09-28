package cmd

import "github.com/spf13/cobra"

func newCheckCmd(rootCmd *cobra.Command) *cobra.Command {
	var checkCmd = &cobra.Command{
		Use:   "check",
		Short: "ark check contains a suite of validation functions",
	}

	rootCmd.AddCommand(checkCmd)
	return checkCmd
}
