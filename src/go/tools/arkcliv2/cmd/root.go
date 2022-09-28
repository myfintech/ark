package cmd

import (
	_ "embed"
	"log"
	"os"
	"runtime/debug"

	"github.com/myfintech/ark/src/go/lib/fs"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/myfintech/ark/src/go/lib/logz"
)

func newRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:           "ark",
		Short:         "ark is the command line interface to interact with the arksdk",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cobra.OnInitialize(newOnInitSetup(rootCmd))
	_ = rootCmd.PersistentFlags().StringP("cwd", "C", "", "sets the current working directory")
	_ = rootCmd.PersistentFlags().String("broker", "nats-embedded", "the message broker to use")
	_ = rootCmd.PersistentFlags().
		String("broker-address", "nats://127.0.0.1:4222", "the address of the message broker")
	_ = rootCmd.PersistentFlags().
		StringP("context", "c", "", "configures the target kubernetes context")
	_ = rootCmd.PersistentFlags().
		StringP("namespace", "n", "", "configures the target kubernetes namespace (does not override static namespaces in manifests)")
	_ = rootCmd.PersistentFlags().
		StringP("environment", "e", "development", "this flag is exposed in the language driver as a CLI flag that can with control logic")
	_ = rootCmd.PersistentFlags().String("log-level", "info", "configures the logging level")
	_ = rootCmd.PersistentFlags().
		String("profile-mode", "", "enables profiling and debugging (cpu|mem|mutex|block)")
	_ = viper.BindPFlags(rootCmd.PersistentFlags())

	// immediately parse persistent flags
	// these are used by later constructors like newLogger
	_ = rootCmd.PersistentFlags().Parse(os.Args[1:])

	return rootCmd
}

type ArkCLI struct {
	RootCmd *cobra.Command
	logger  *logz.Writer
}

func (a ArkCLI) Execute() (exitCode int) {
	defer func() {
		a.logger.Close()
		_ = a.logger.Wait()
	}()

	defer func() {
		if v := recover(); v != nil {
			exitCode = 1
			a.logger.Errorf("%v %s", v, debug.Stack())
		}
	}()

	if err := a.RootCmd.Execute(); err != nil {
		exitCode = 1
		a.logger.Error(err)
		return
	}

	return
}

func newArkCLI(rootCmd *cobra.Command, core coreConfig) (*ArkCLI, error) {
	arkCLI := &ArkCLI{
		RootCmd: rootCmd,
		logger:  core.logger,
	}

	checkCmd := newCheckCmd(rootCmd)
	newCheckGlobCmd(checkCmd, core.logger, core.config, core.gitIgnorePatterns)
	newCheckIgnoreCmd(checkCmd, core.config, core.fileObserver, core.logger)
	newInitCmd(rootCmd, core.config)

	kvCmd := newKVCmd(rootCmd)
	newKVDeleteCmd(kvCmd, core.kvStorage, core.config)
	newKVEditCmd(kvCmd, core.kvStorage, core.config)
	newKVGetCmd(kvCmd, core.kvStorage, core.config)
	newKVImportCmd(kvCmd, core.vaultClient, core.kvStorage, core.config)
	newKVPutCmd(kvCmd, core.config, core.kvStorage)

	serverCmd := newServerCmd(rootCmd)
	newServerRestartCmd(serverCmd, core.logger, core.hostServerDaemon)
	newServerRunCmd(
		serverCmd,
		core.fileObserver,
		core.subsystemsManager,
		core.sharedClients,
		core.store,
		core.logger,
		core.config,
	)
	newServerStarCmd(serverCmd, core.hostServerDaemon)
	newServerStatusCmd(serverCmd, core.logger, core.hostServerDaemon)
	newServerStopCmd(serverCmd, core.hostServerDaemon)

	targetsCmd := newTargetsCmd(rootCmd)
	newTargetsListCmd(targetsCmd, core.httpClient)

	newInitCmd(rootCmd, core.config)
	newVersionCmd(rootCmd, core.logger)
	newUpgradeCmd(rootCmd, core.logger)
	newLogsCmd(rootCmd, core.httpClient)
	newRunCmd(rootCmd, core.logger, core.config, core.vm, core.httpClient, core.hostServerDaemon)

	return arkCLI, nil
}

// initConfig reads in config file and ENV variables if set.
func newOnInitSetup(rootCmd *cobra.Command) func() {
	return func() {

		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalln(err)
		}

		userWD, err := rootCmd.Flags().GetString("cwd")
		if err != nil {
			log.Fatalln(err)
		}

		if userWD != "" {
			userWD, err = fs.NormalizePath(cwd, userWD)
			if err != nil {
				log.Fatalln(err)
			}
			if err = os.Chdir(userWD); err != nil {
				log.Fatalln(errors.Wrap(err, "failed to change to specified cwd"))
			}
		}

		// sets PROGRESS_NO_TRUNC env var to prevent plaintext Docker logs from being truncated
		if err = os.Setenv("PROGRESS_NO_TRUNC", "1"); err != nil {
			log.Fatalln(err)
		}

		viper.SetEnvPrefix("ARK")
		// read in environment variables that match
		viper.AutomaticEnv()
	}

}
