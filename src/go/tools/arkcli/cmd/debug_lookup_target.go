package cmd

import (
	"fmt"

	"github.com/myfintech/ark/src/go/lib/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var debugLookupTargetCmd = &cobra.Command{
	Use:   "lookup_target",
	Short: "lookup_target is a sub-command of debug that outputs debug info on the specified object",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print(utils.MarshalJSONSafe(workspace.TargetLUT, true))
		return nil
	},
}

func init() {
	debugCmd.AddCommand(debugLookupTargetCmd)
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
}
