package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/myfintech/ark/src/go/lib/log"

	"github.com/spf13/viper"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
)

var workspace = base.NewWorkspace()

// rootCmd represents the base command when called without any sub-commands
var rootCmd = &cobra.Command{
	Use:           "ark",
	Short:         "ark is the command line interface to interact with the arksdk",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func RootCmd() *cobra.Command {
	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	defer func() {
		if err := recover(); err != nil {
			if viper.GetBool("debug") {
				panic(err)
			}
			fmt.Println(err)
			os.Exit(1)
		}
	}()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().String("log_level", "info", "Logging level. Supports 'trace', 'debug', 'info', 'warn', 'error', and 'fatal'")
	rootCmd.PersistentFlags().String("log_format", "text", "The log output format. Supports 'text' and 'json'")
	rootCmd.PersistentFlags().Bool("skip_version_check", false, "Skip the version check when calling ark")
	rootCmd.PersistentFlags().Bool("debug", false, "Captures a more detailed stack trace on error if there was a panic")

	rootCmd.PersistentFlags().StringP("namespace", "n", "", "Defines what namespace should be used for Kubernetes actions")
	rootCmd.PersistentFlags().String("environment", "", "Defines what environment should be considered for configuration loading")
	rootCmd.PersistentFlags().String("cpu_profile", "", "Runs benchmarks against cpu usage")
	rootCmd.PersistentFlags().String("mem_profile", "", "Runs benchmarks against memory usage")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	_ = viper.BindPFlags(rootCmd.PersistentFlags())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetEnvPrefix("")
	viper.AutomaticEnv() // read in environment variables that match
	log.Init()
}
