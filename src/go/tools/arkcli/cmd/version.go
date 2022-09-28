package cmd

import (
	"fmt"

	"github.com/myfintech/ark/src/go/lib/pkg"
	"github.com/myfintech/ark/src/go/lib/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "version is a sub-command of ark that provides version information about the arkcli",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(utils.MarshalJSONSafe(pkg.GlobalInfo(), true))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
}
