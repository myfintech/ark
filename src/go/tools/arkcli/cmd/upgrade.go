package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/myfintech/ark/src/go/lib/log"
	"github.com/myfintech/ark/src/go/lib/pkg"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "executes an ark upgrade",
	RunE: func(cmd *cobra.Command, args []string) error {
		return pkg.VersionCheckHook(false, func(current, latest pkg.PackageInfo) error {
			log.Infof("Upgrading from %s to %s", current.Version, latest.Version)
			if err := pkg.Upgrade(latest); err != nil {
				return errors.Wrapf(err, "the upgrade was unsuccessful; rolling back upgrade")
			}
			return nil
		}, func(_ pkg.PackageInfo, _ pkg.PackageInfo) error {
			fmt.Print("The arkcli is up to date at this time.\n")
			return nil
		})
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
}
