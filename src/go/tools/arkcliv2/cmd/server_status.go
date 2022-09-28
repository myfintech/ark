package cmd

import (
	"github.com/myfintech/ark/src/go/lib/daemonize"
	"github.com/myfintech/ark/src/go/lib/logz"
	"github.com/spf13/cobra"
)

func newServerStatusCmd(
	serverCmd *cobra.Command,
	logger logz.FieldLogger,
	daemon *daemonize.Proc,
) *cobra.Command {
	var serverStatusCmd = &cobra.Command{
		Use:  "status",
		Long: "ark server status queries your host server daemon manager for the status of the ark host server",
		RunE: func(cmd *cobra.Command, args []string) error {
			state, err := daemon.Status()
			switch state {
			case daemonize.STATE_RUNNING:
				logger.Infof("running")
			case daemonize.STATE_STOPPED:
				logger.Infof("stopped")
			case daemonize.STATE_UNKNOWN:
				logger.Infof("unknown")
			}
			return err
		},
	}

	serverCmd.AddCommand(serverStatusCmd)
	return serverStatusCmd
}
