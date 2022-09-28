package cmd

import (
	"github.com/myfintech/ark/src/go/lib/daemonize"
	"github.com/spf13/cobra"
)

func newServerStarCmd(
	serverCmd *cobra.Command,
	daemon *daemonize.Proc,
) *cobra.Command {
	var serverStartCmd = &cobra.Command{
		Use:  "start",
		Long: "ark server start tells the system daemon manager to start the ark host server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return daemon.Init()
		},
	}

	serverCmd.AddCommand(serverStartCmd)
	return serverStartCmd
}
