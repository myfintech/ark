package cmd

import (
	"github.com/myfintech/ark/src/go/lib/ark/kv"
	"github.com/myfintech/ark/src/go/lib/ark/workspace"
	"github.com/spf13/cobra"
)

func newKVEditCmd(kvCmd *cobra.Command, storage kv.Storage, config *workspace.Config) *cobra.Command {
	var kvEditCmd = &cobra.Command{
		Use:               "edit SECRET_PATH",
		Short:             "edit is a sub-command of kv that reads configuration into a temp file and opens the temp file in the user's $EDITOR for modification",
		PreRunE:           requireKVSecretPath(config),
		ValidArgsFunction: getValidKVFiles,
		RunE: func(cmd *cobra.Command, args []string) error {
			return storage.Edit(args[0])
		},
	}

	kvCmd.AddCommand(kvEditCmd)
	return kvEditCmd
}
