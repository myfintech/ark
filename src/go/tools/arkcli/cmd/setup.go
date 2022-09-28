package cmd

import (
	"encoding/json"
	"io/ioutil"
	"runtime"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/myfintech/ark/src/go/lib/log"

	"github.com/myfintech/ark/src/go/lib/fs"
)

var macosDockerConfigPath = "~/Library/Group Containers/group.com.docker/settings.json"

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "configures your system to use ark",
	RunE: func(cmd *cobra.Command, args []string) error {
		if runtime.GOOS != "darwin" {
			log.Infof("there are no setup steps for %s", runtime.GOOS)
			return nil
		}

		path, err := homedir.Expand(macosDockerConfigPath)
		if err != nil {
			return err
		}

		data, err := fs.ReadFileBytes(path)
		if err != nil {
			return err
		}

		var rawDockerConfig map[string]interface{}
		var dockerConfig struct {
			FileSharingDirectories []string `json:"FileSharingDirectories"`
		}

		err = json.Unmarshal(data, &rawDockerConfig)
		if err != nil {
			return err
		}

		err = json.Unmarshal(data, &dockerConfig)
		if err != nil {
			return errors.Wrapf(err, "did not recognize docker config at %s", macosDockerConfigPath)
		}

		rawDockerConfig["kubernetesEnabled"] = true

		updatedConfig, err := json.MarshalIndent(rawDockerConfig, "", "  ")
		if err != nil {
			return errors.Wrap(err, "failed to marshal updated docker config to JSON")
		}

		err = ioutil.WriteFile(path, updatedConfig, 0644)
		if err != nil {
			return err
		}

		log.Info("Your docker configuration has been updated, please restart docker.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
}
