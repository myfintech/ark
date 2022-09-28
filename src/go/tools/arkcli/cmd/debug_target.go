package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/myfintech/ark/src/go/lib/utils"
)

const debugTargetDescription = `
target is a sub-command of debug that outputs target details
`

var debugTargetCmd = &cobra.Command{
	Use:               "target TARGET_ADDRESS",
	Short:             debugTargetDescription,
	ValidArgsFunction: getValidTargets,
	RunE: func(cmd *cobra.Command, args []string) error {

		fmCache, _ := workspace.Observer.GetMatchCache(args[0])
		if fmCache == nil {
			return errors.Errorf("%s is not a valid target address", args[0])
		}

		fmt.Println(utils.MarshalJSONSafe(map[string]interface{}{
			"source_files": fmCache.SortedFilesStringList(),
		}, true))

		return nil
	},
}

func init() {
	debugCmd.AddCommand(debugTargetCmd)
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
}
