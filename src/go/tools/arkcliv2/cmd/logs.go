package cmd

import (
	"io"
	"os"

	"github.com/myfintech/ark/src/go/lib/ark/subsystems/http_server"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

func newLogsCmd(
	rootCmd *cobra.Command,
	client http_server.Client,
) *cobra.Command {
	var logsCmd = &cobra.Command{
		Use:     "logs {all | SUBSCRIPTION_ID}",
		Short:   "logs is a sub-command of ark that exposes commands for streaming logs",
		Aliases: []string{"log"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("a parameter of 'all' or a subscription ID must be provided to the logs command")
			}

			logID := args[0]

			var err error
			var logStream io.Reader

			switch logID {
			case "all":
				logStream, err = client.GetServerLogs()
			default:
				logStream, err = client.GetLogsByKey(logID)
			}

			if err != nil {
				return err
			}

			if _, err = io.Copy(os.Stdout, logStream); err != nil {
				return err
			}

			return nil
		},
	}

	rootCmd.AddCommand(logsCmd)
	return logsCmd
}
