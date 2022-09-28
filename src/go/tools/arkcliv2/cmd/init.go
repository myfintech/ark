package cmd

import (
	"github.com/myfintech/ark/src/go/lib/ark/embeds"
	"github.com/myfintech/ark/src/go/lib/ark/workspace"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newInitCmd(
	rootCmd *cobra.Command,
	config *workspace.Config,
) *cobra.Command {
	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "init is a sub-command of ark that initializes ...",
		RunE: func(cmd *cobra.Command, args []string) error {
			return embeds.Unpack(config.Root())
		},
	}

	rootCmd.AddCommand(initCmd)
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
	return initCmd
}
