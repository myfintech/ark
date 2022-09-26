package cmd

import (
	"github.com/myfintech/ark/src/go/lib/daemonize"
	"github.com/myfintech/ark/src/go/lib/logz"
	"github.com/spf13/cobra"
)

func newServerRestartCmd(
	serverCmd *cobra.Command,
	logger logz.FieldLogger,
	hostServerDaemon *daemonize.Proc,
) *cobra.Command {
	var serverRestartCmd = &cobra.Command{
		Use:  "restart",
		Long: "ark server restart stops and starts the host daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := hostServerDaemon.Stop(); err != nil {
				logger.Debugf("failed to stop %v", err)
			}
			return hostServerDaemon.Init()
		},
	}

	serverCmd.AddCommand(serverRestartCmd)
	return serverRestartCmd
}
