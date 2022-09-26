package cmd

import (
	"os"
	"path/filepath"

	"github.com/myfintech/ark/src/go/lib/ark/workspace"

	"github.com/myfintech/ark/src/go/lib/ark/kv"

	"github.com/spf13/cobra"
)

func newKVDeleteCmd(kvCmd *cobra.Command, kvStorage kv.Storage, config *workspace.Config) *cobra.Command {
	var kvDeleteCmd = &cobra.Command{
		Use:               "delete SECRET_PATH",
		Short:             "delete is a sub-command of kv that deletes a kv file",
		PreRunE:           requireKVSecretPath(config),
		ValidArgsFunction: getValidKVFiles,
		RunE: func(cmd *cobra.Command, args []string) error {
			return os.Remove(filepath.Join(kvStorage.EncryptedDataPath(), args[0]))
		},
	}

	kvCmd.AddCommand(kvDeleteCmd)
	return kvDeleteCmd
}
