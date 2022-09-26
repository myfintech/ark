package cmd

import (
	"github.com/cheynewallace/tabby"
	"github.com/myfintech/ark/src/go/lib/pattern"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
)

var targetsListCmd = &cobra.Command{
	Use:   "list",
	Short: "list is a sub-command of targets that lists all buildable targets",
	RunE: func(cmd *cobra.Command, args []string) error {
		var matcher *pattern.Matcher
		t := tabby.New()
		t.AddHeader("short_hash", "target_address", "locally_cached", "remotely_cached", "labels", "description")

		remote, err := cmd.Flags().GetBool("remote")
		if err != nil {
			return err
		}

		patterns, err := cmd.Flags().GetStringSlice("filters")
		if err != nil {
			return err
		}

		if len(patterns) > 0 {
			matcher = &pattern.Matcher{
				Includes: patterns,
			}
			if err = matcher.Compile(); err != nil {
				return err
			}
		}

		for _, address := range workspace.TargetLUT.FilterSortedAddresses(matcher) {
			addressable := workspace.TargetLUT[address]
			buildableTarget, buildable := addressable.(base.Buildable)
			cacheableTarget, cacheable := addressable.(base.Cacheable)
			locallyCached := false
			remotelyCached := false

			if buildable {
				if err = buildableTarget.PreBuild(); err != nil {
					return err
				}
			}
			if cacheable {
				locallyCached, err = cacheableTarget.CheckLocalBuildCache()
				if err != nil {
					return err
				}

				if remote {
					remotelyCached, err = cacheableTarget.CheckRemoteCache()
					if err != nil {
						return err
					}
				}
			}
			t.AddLine(buildableTarget.ShortHash(), address, locallyCached, remotelyCached, buildableTarget.ListLabels(), addressable.Describe())
		}
		t.Print()
		return nil
	},
}

func init() {
	targetsCmd.AddCommand(targetsListCmd)
	_ = targetsListCmd.Flags().Bool("remote", false, "check remote cache")
	_ = targetsListCmd.Flags().StringSlice("filters", nil, "filters used to limit list results")
	_ = viper.BindPFlags(targetsCmd.PersistentFlags())
}
