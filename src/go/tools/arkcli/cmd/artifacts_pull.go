package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/myfintech/ark/src/go/lib/log"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
)

var artifactsPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "pull is a sub-command of artifacts that pulls buildable artifacts from a remote cache",
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, addressable := range workspace.TargetLUT {
			if target, cacheable := addressable.(base.Cacheable); cacheable {
				if cached, _ := target.CheckRemoteCache(); cached {
					if err := target.PullRemoteCache(); err != nil {
						return err
					}
					log.Infof("pulled artifact for %s target", addressable.Address())
				} else {
					log.Debugf("no remote cache for %s target; skipping ... ", addressable.Address())
				}
			}
		}
		return nil
	},
}

func init() {
	artifactsCmd.AddCommand(artifactsPullCmd)
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
}
