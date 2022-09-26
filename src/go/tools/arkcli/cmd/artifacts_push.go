package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/myfintech/ark/src/go/lib/log"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
)

var artifactsPushCmd = &cobra.Command{
	Use:   "push",
	Short: "push is a sub-command of artifacts that pushes buildable artifacts to a remote cache",
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, addressable := range workspace.TargetLUT {
			buildableTarget, buildable := addressable.(base.Buildable)
			cacheableTarget, cacheable := addressable.(base.Cacheable)
			if buildable {
				if err := buildableTarget.PreBuild(); err != nil {
					return err
				}
			}

			if cacheable {
				if cached, _ := cacheableTarget.CheckLocalBuildCache(); cached {
					if err := cacheableTarget.PushRemoteCache(); err != nil {
						return err
					}
					log.Infof("pushed artifact for %s target", addressable.Address())
				} else {
					log.Debugf("no local cache for %s target; skipping ... ", addressable.Address())
				}
			}
		}
		return nil
	},
}

func init() {
	artifactsCmd.AddCommand(artifactsPushCmd)
	// FIXME: remove rootCmd persistent flags from sub-commands to prevent overwriting root-level flags
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
}
