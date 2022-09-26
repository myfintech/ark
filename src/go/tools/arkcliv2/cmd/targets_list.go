package cmd

import (
	"github.com/cheynewallace/tabby"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems/http_server"
	"github.com/spf13/cobra"
)

func newTargetsListCmd(
	targetsCmd *cobra.Command,
	httpClient http_server.Client,
) *cobra.Command {
	var targetsListCmd = &cobra.Command{
		Use:   "list",
		Short: "list is a sub-command of targets that lists all buildable targets",
		RunE: func(cmd *cobra.Command, args []string) error {
			t := tabby.New()
			t.AddHeader("target_key")

			targets, err := httpClient.GetTargets()
			if err != nil {
				return err
			}

			for _, target := range targets {
				t.AddLine(target.Key())
			}

			t.Print()
			return nil
		},
	}

	targetsCmd.AddCommand(targetsListCmd)
	return targetsListCmd
}
