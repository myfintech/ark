package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/myfintech/ark/src/go/lib/autoupdate"
	"github.com/spf13/cobra"
)

func newVersionCmd(
	rootCmd *cobra.Command,
	logger logz.FieldLogger,
) *cobra.Command {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "version check the current version",
		RunE: func(cmd *cobra.Command, args []string) error {
			updatable, remoteVersion, err := autoupdate.CheckVersion()
			if err != nil {
				logger.Warnf("failed to check remote version %v", err)
			}

			data, err := json.MarshalIndent(map[string]interface{}{
				"local":  autoupdate.CurrentVersion(),
				"remote": remoteVersion,
			}, "", "  ")

			if err != nil {
				return err
			}

			fmt.Println(string(data))

			logger.Infof("OS: %s Arch: %s", runtime.GOOS, runtime.GOARCH)

			if updatable {
				logger.Info("A new version is available")
			} else {
				logger.Info("You're running the latest version")
			}

			return nil
		},
	}

	rootCmd.AddCommand(versionCmd)
	return versionCmd
}
