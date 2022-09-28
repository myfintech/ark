package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/myfintech/ark/src/go/lib/fs"
	"github.com/myfintech/ark/src/go/lib/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var kvGetCmd = &cobra.Command{
	Use:   "get SECRET_PATH",
	Short: "get is a sub-command of kv that reads configuration and prints the contents",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("SECRET_PATH is required")
		}

		if workspace.Config.Vault == nil || workspace.Config.Vault.EncryptionKey == "" {
			return errors.New("must specify vault.encryption_key in WORKSPACE.hcl")
		}
		return nil
	},
	ValidArgsFunction: getValidKVFiles,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		config, err := workspace.KVStorage.Get(path)
		if err != nil {
			return err
		}

		fmt.Println(utils.MarshalJSONSafe(config, true))

		return nil
	},
}

func getValidKVFiles(_ *cobra.Command, _ []string, complete string) ([]string, cobra.ShellCompDirective) {
	completions := make([]string, 0)
	arkKVPath := filepath.Join(workspace.Dir, ".ark", "kv")
	secretPath := filepath.Join(arkKVPath, complete)
	_ = filepath.Walk(secretPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() == false && strings.HasPrefix(path, secretPath) {
			completions = append(completions, fs.TrimPrefix(path, arkKVPath))
		}
		return nil
	})
	return completions, cobra.ShellCompDirectiveNoFileComp
}

func init() {
	kvCmd.AddCommand(kvGetCmd)
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
}
