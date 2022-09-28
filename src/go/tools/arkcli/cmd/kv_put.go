package cmd

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/myfintech/ark/src/go/lib/fs"
)

var kvPutCmd = &cobra.Command{
	Use:   "put SECRET_PATH",
	Short: "put is a sub-command of kv that writes configuration key/pair values to local Vault",
	Long: `
ark kv put secrets/foo bar=baz qux=quux
ark kv put secrets/foo --file ./example.json
cat example.json | ark kv put secrets/foo -f -
`,
	ValidArgsFunction: getValidKVFiles,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("SECRET_PATH is a required parameter")
		}

		file, _ := cmd.Flags().GetString("file")
		if len(args) < 2 && file == "" {
			return errors.New("either kv pairs (key=value) or --file must be specified")
		}

		if workspace.Config.Vault == nil || workspace.Config.Vault.EncryptionKey == "" {
			return errors.New("must specify vault.encryption_key in WORKSPACE.hcl")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		file, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}

		path := args[0]
		config := make(map[string]interface{})

		for _, p := range args[1:] {
			pairs := strings.Split(p, "=")
			if len(pairs) != 2 {
				return errors.New("key value pairs must be in format key=value")
			}
			config[pairs[0]] = pairs[1]
		}

		switch {
		case file == "-":
			decoder := json.NewDecoder(os.Stdin)
			if dErr := decoder.Decode(&config); dErr != nil {
				return dErr
			}
			break
		case file != "":
			cwd, fErr := os.Getwd()
			if fErr != nil {
				return fErr
			}

			file, fErr = fs.NormalizePath(cwd, file)
			if fErr != nil {
				return fErr
			}

			if fErr = fs.ReadFileJSON(file, &config); fErr != nil {
				return fErr
			}
			break
		}

		if file != "" && file != "-" {
		}

		_, err = workspace.KVStorage.Put(path, config)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	kvCmd.AddCommand(kvPutCmd)
	kvPutCmd.Flags().StringP("file", "f", "", "The path to a JSON encoded file to import into the given path")
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
}
