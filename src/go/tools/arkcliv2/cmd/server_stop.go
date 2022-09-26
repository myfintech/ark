package cmd

import (
	"github.com/myfintech/ark/src/go/lib/daemonize"
	"github.com/spf13/cobra"
)

func newServerStopCmd(
	serverCmd *cobra.Command,
	daemon *daemonize.Proc,
) *cobra.Command {
	var serverStopCmd = &cobra.Command{
		Use:  "stop",
		Long: "ark server stop tells the system daemon manager to stop the ark host server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return daemon.Stop()
		},
	}

	serverCmd.AddCommand(serverStopCmd)
	return serverStopCmd
}
