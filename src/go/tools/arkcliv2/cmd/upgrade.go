package cmd

import (
	"os"

	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/myfintech/ark/src/go/lib/autoupdate"
	"github.com/spf13/cobra"
)

func newUpgradeCmd(
	rootCmd *cobra.Command,
	logger logz.FieldLogger,
) *cobra.Command {
	var upgradeCmd = &cobra.Command{
		Use:   "upgrade",
		Short: "upgrade is a sub-command of ark that upgrades the Ark binary in place",
		RunE: func(cmd *cobra.Command, args []string) error {
			updatable, remote, err := autoupdate.CheckVersion()
			if err != nil {
				return err
			}

			if !updatable {
				logger.Info("You're all up to date")
				return nil
			}

			executablePath, err := os.Executable()
			if err != nil {
				return err
			}

			logger.Infof("Preparing to install version %s", remote.Version)
			if err = autoupdate.Upgrade(executablePath); err != nil {
				return err
			}

			logger.Info("upgrade successful")

			return nil
		},
	}

	rootCmd.AddCommand(upgradeCmd)
	return upgradeCmd
}
