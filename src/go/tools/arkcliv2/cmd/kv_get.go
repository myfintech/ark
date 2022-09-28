package cmd

import (
	"fmt"

	"github.com/myfintech/ark/src/go/lib/ark/workspace"

	"github.com/myfintech/ark/src/go/lib/ark/kv"

	"github.com/myfintech/ark/src/go/lib/utils"
	"github.com/spf13/cobra"
)

func newKVGetCmd(kvCmd *cobra.Command, storage kv.Storage, config *workspace.Config) *cobra.Command {
	var kvGetCmd = &cobra.Command{
		Use:               "get SECRET_PATH",
		Short:             "get is a sub-command of kv that reads configuration and prints the contents",
		PreRunE:           requireKVSecretPath(config),
		ValidArgsFunction: getValidKVFiles,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			config, err := storage.Get(path)
			if err != nil {
				return err
			}

			fmt.Println(utils.MarshalJSONSafe(config, true))

			return nil
		},
	}

	kvCmd.AddCommand(kvGetCmd)
	return kvGetCmd

}
