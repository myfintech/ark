package cmd

import (
	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Cleans state of local workspace",
	Long:  `Removes all files and folders in the /ark store`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return workspace.Clean()
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
