package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/cli/cli/config"
	"github.com/myfintech/ark/src/go/lib/ark/workspace"
	"github.com/myfintech/ark/src/go/lib/fs"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newKVCmd(rootCmd *cobra.Command) *cobra.Command {
	var kvCmd = &cobra.Command{
		Use:   "kv",
		Short: "kv is a sub-command of ark that manages configurations for local Vault",
	}

	rootCmd.AddCommand(kvCmd)
	_ = viper.BindPFlags(kvCmd.PersistentFlags())
	return kvCmd
}

func requireKVSecretPath(config *workspace.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("SECRET_PATH is required")
		}

		if config.Vault.EncryptionKey == "" {
			return errors.New("must specify vault.encryption_key in workspace settings")
		}
		return nil
	}
}

func getValidKVFiles(_ *cobra.Command, _ []string, complete string) ([]string, cobra.ShellCompDirective) {
	completions := make([]string, 0)
	arkKVPath := filepath.Join(config.Dir(), ".ark", "kv")
	secretPath := filepath.Join(arkKVPath, complete)
	_ = filepath.Walk(secretPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasPrefix(path, secretPath) {
			completions = append(completions, fs.TrimPrefix(path, arkKVPath))
		}
		return nil
	})
	return completions, cobra.ShellCompDirectiveNoFileComp
}
