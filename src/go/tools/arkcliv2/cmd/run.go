package cmd

import (
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/protocols/nats"
	nats2 "github.com/nats-io/nats.go"

	"github.com/myfintech/ark/src/go/lib/ark/subsystems/http_server"

	"github.com/myfintech/ark/src/go/lib/ark/workspace"

	"github.com/myfintech/ark/src/go/lib/kube"

	"golang.org/x/term"

	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/commands"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"

	"github.com/myfintech/ark/src/go/lib/daemonize"

	"github.com/myfintech/ark/src/go/lib/internal_net"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/events"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript"

	tea "github.com/charmbracelet/bubbletea"

	"golang.org/x/sync/errgroup"

	"github.com/myfintech/ark/src/go/tools/arkcliv2/cmd/ui/graph_progress"

	fs2 "github.com/myfintech/ark/src/go/lib/fs"

	"github.com/myfintech/ark/src/go/lib/ark"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/messages"

	"github.com/moby/buildkit/util/appcontext"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var natsDisconnected = make(chan error, 1)

func logBuildLogURL(logger logz.FieldLogger, buildId string) {
	logger.Infof("$$$ ==> ||  Run in terminal: ark logs %s", buildId)
	logger.Infof("*** ==> ||  Build Log URL: http://127.0.0.1:9000/server/logs/%s", buildId)
	if path, err := logz.SuggestedFilePath("ark/graph", buildId); err == nil {
		logger.Infof("### ==> ||  logs are located at: %s", path)
	}

}

func newRunCmd(
	rootCmd *cobra.Command,
	logger logz.FieldLogger,
	config *workspace.Config,
	vm *typescript.VirtualMachine,
	serverClient http_server.Client,
	hostServerDaemon *daemonize.Proc,
) *cobra.Command {
	var runCmd = &cobra.Command{
		Use:     "run TARGET_PATH TARGET_NAME",
		Short:   "run executes the target entrypoint",
		Long:    `ark run src/go/services/my_service/build.ts goModules`,
		PreRunE: validateArgsRequired,
		Args:    cobra.MinimumNArgs(2),
		PersistentPreRunE: cobraRunEMiddleware(
			ensureServerRunning(hostServerDaemon, logger),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()
			targetPath := args[0]
			targetName := args[1]

			defer func() { logger.Infof("executed in %s", time.Now().Sub(start)) }()

			disconnectErrChan := make(chan error)
			conn, err := nats.Connect("nats://127.0.0.1:4222", func(options *nats2.Options) error {
				options.DisconnectedErrCB = func(conn *nats2.Conn, err error) {
					disconnectErrChan <- errors.Wrap(err, "server broker disconnected")
				}
				return nil
			})
			if err != nil {
				return err
			}

			broker := nats.NewBroker(conn)

			cwd, err := os.Getwd()
			if err != nil {
				return nil
			}
			if !filepath.IsAbs(targetPath) {
				targetPath, err = fs2.NormalizePath(cwd, targetPath)
				if err != nil {
					return err
				}
			}

			logger.Infof("entrypoint %s", targetPath)

			async, err := cmd.Flags().GetBool("async")
			if err != nil {
				return err
			}

			k8sNamespace, err := cmd.Flags().GetString("namespace")
			if err != nil {
				return err
			}

			environment, err := cmd.Flags().GetString("environment")
			if err != nil {
				return err
			}

			maxConcurrency, err := cmd.Flags().GetInt("max-concurrency")
			if err != nil {
				return err
			}

			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return err
			}

			k8sContext, err := cmd.Flags().GetString("context")
			if err != nil {
				return err
			}

			push, err := cmd.Flags().GetBool("push")
			if err != nil {
				return err
			}

			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return err
			}

			skip, err := cmd.Flags().GetStringSlice("skip")
			if err != nil {
				return err
			}

			ciMode, err := cmd.Flags().GetBool("ci")
			if err != nil {
				return err
			}

			if k8sContext != "" {
				panic("--context is not implemented")
			}

			if k8sNamespace == "" {
				k8sNamespace = config.K8s.Namespace
			}

			k8sNamespace = kube.NormalizeNamespace(k8sNamespace)

			err = InstallCLIModules(vm, cmd, args, map[string]interface{}{
				"namespace":   k8sNamespace,
				"context":     k8sContext,
				"environment": environment,
				"ci":          ciMode,
			})
			if err != nil {
				return err
			}

			target := ark.RawTarget{
				File:  targetPath,
				Name:  targetName,
				Realm: config.Root(),
			}

			logger.Infof("resolving workspace build files (this could take some time)")
			if _, err = vm.ResolveModule(targetPath); err != nil {
				return err
			}
			logger.Infof("resolved in %s", time.Now().Sub(start))

			if dryRun {
				return nil
			}

			r, err := serverClient.Run(messages.GraphRunnerExecuteCommand{
				TargetKeys:     []string{target.Key()},
				K8sNamespace:   k8sNamespace,
				PushAfterBuild: push,
				ForceBuild:     force,
				SkipFilters:    skip,
				MaxConcurrency: maxConcurrency,
				K8sContext:     k8sContext,
			})

			if err != nil {
				return err
			}

			logger.Infof("build ID %s", r.SubscriptionId)
			logBuildLogURL(logger, r.SubscriptionId)

			if async {
				return nil
			}

			// subscribe to all events and filter events by subject
			stream, err := broker.Subscribe(appcontext.Context(), ">", nil)
			if err != nil {
				return err
			}

			eg, _ := errgroup.WithContext(appcontext.Context())

			if term.IsTerminal(int(os.Stdout.Fd())) && !ciMode {
				interactiveTUIMode(eg, stream, r, logger, broker)
			} else {
				fallbackRawOutputMode(eg, stream, r, logger, broker)
			}

			return eg.Wait()
		},
	}

	rootCmd.AddCommand(runCmd)
	_ = runCmd.PersistentFlags().Bool("ci", false, "use this flag to disable interactive mode")
	_ = runCmd.PersistentFlags().Bool("dry-run", false, "use this flag to load the action graph without execution")
	_ = runCmd.PersistentFlags().Bool("force", false, "ignores cache and forces action action execution")
	_ = runCmd.PersistentFlags().Bool("push", false, "pushes artifacts after successful actions (use for incremental CI builds)")
	_ = runCmd.PersistentFlags().Bool("async", false, "returns the subscription id of the graph run to resume watching later")
	_ = runCmd.PersistentFlags().StringSlice("skip", []string{}, "[DONT USE] supplies patterns to skip actions in the graph (useful for skipping tests, can cause unexpected behavior)")
	_ = runCmd.PersistentFlags().IntP("max-concurrency", "m", runtime.GOMAXPROCS(0), "Sets a limit on the graph walk parallelism [default based on available CPUs]")

	return runCmd
}

func fallbackRawOutputMode(
	eg *errgroup.Group,
	stream <-chan cqrs.Envelope,
	r messages.GraphRunnerExecuteCommandResponse,
	logger logz.FieldLogger,
	broker cqrs.Broker,
) {
	eg.Go(func() error {
		for {
			select {
			case err := <-natsDisconnected:
				return err
			case envelope := <-stream:
				if envelope.Subject() != r.SubscriptionId {
					continue
				}

				switch envelope.TypeKey() {
				case events.GraphRunnerFailed:
					return errors.Errorf("graph runner failed: %s", string(envelope.Data()))
				case events.GraphRunnerSuccess:
					return nil
				}

				var d ark.Derivative
				if err := envelope.DataAs(&d); err != nil {
					continue
				}

				logger.WithFields(logz.Fields{
					"hash":   d.RawArtifact.ShortHash(),
					"target": d.RawTarget.Key(),
					"event":  envelope.Type(),
				}).Info()
			case <-appcontext.Context().Done():
				logger.Infof("sending cancellation signal for %s", r.SubscriptionId)
				pubErr := broker.Publish(topics.GraphRunnerCommands, cqrs.NewDefaultEnvelope(
					commands.GraphRunnerCancelType,
					cqrs.WithSource("arkcli"),
					cqrs.WithSubject(cqrs.RouteKey(r.SubscriptionId)),
				))
				if pubErr != nil {
					return errors.Wrap(pubErr, "failed to publish cancellation")
				}
				return appcontext.Context().Err()
			}
		}
	})
}

func interactiveTUIMode(
	eg *errgroup.Group,
	stream <-chan cqrs.Envelope,
	r messages.GraphRunnerExecuteCommandResponse,
	logger logz.FieldLogger,
	broker cqrs.Broker,
) {
	uiStream := make(chan tea.Msg, 1000)

	eg.Go(func() error {
		for {
			select {
			case err := <-natsDisconnected:
				uiStream <- graph_progress.Stop()
				return err
			case envelope := <-stream:
				if envelope.Subject() != r.SubscriptionId {
					continue
				}

				uiStream <- envelope
				switch envelope.TypeKey() {
				case events.GraphRunnerFailed:
					uiStream <- graph_progress.Stop()
					logBuildLogURL(logger, r.SubscriptionId)

					return errors.Errorf("graph runner failed: %s", string(envelope.Data()))
				case events.GraphRunnerSuccess:
					uiStream <- graph_progress.Stop()
					return nil
				}
			case <-appcontext.Context().Done():
				logger.Infof("sending cancellation signal for %s", r.SubscriptionId)
				pubErr := broker.Publish(topics.GraphRunnerCommands, cqrs.NewDefaultEnvelope(
					commands.GraphRunnerCancelType,
					cqrs.WithSource("arkcli"),
					cqrs.WithSubject(cqrs.RouteKey(r.SubscriptionId)),
				))
				if pubErr != nil {
					return errors.Wrap(pubErr, "failed to publish cancellation")
				}
				select {
				case uiStream <- graph_progress.Stop():
					return errors.New("canceled")
				case <-time.After(time.Second * 5):
					return errors.New("failed gracefully stop terminal UI")
				}
			}
		}
	})
	eg.Go(graph_progress.New(uiStream))
}

func ensureServerRunning(
	daemon *daemonize.Proc,
	logger logz.FieldLogger,
) runE {
	return func(cmd *cobra.Command, args []string) error {
		err := daemon.Init()
		if err != nil && !errors.Is(err, daemonize.ErrAlreadyRunning) {
			return err
		}

		host, err := url.Parse("http://127.0.0.1:9000/health")
		if err != nil {
			return err
		}

		probe, options, err := internal_net.CreateProbe(internal_net.ProbeOptions{
			Timeout:    time.Second * 2,
			Delay:      time.Second * 3,
			Address:    host,
			MaxRetries: 2,
			OnError: func(err error, remainingAttempts int) {
				logger.Debugf("an error occurred while waiting for host server to start %v", err)
			},
		})

		if err != nil {
			return err
		}

		if err = internal_net.RunProbe(options, probe); err != nil {
			return err
		}
		return nil
	}
}

// CMD an interface that mimics cobra to define dependencies required by InstallCLIModules
type CMD interface {
	ArgsLenAtDash() int
}

// InstallCLIModules forwards OS and CLI arguments to typescript space
func InstallCLIModules(vm *typescript.VirtualMachine, cmd CMD, args []string, flags map[string]interface{}) error {
	var extraCLIArgs []string
	if cmd.ArgsLenAtDash() != -1 {
		extraCLIArgs = args[cmd.ArgsLenAtDash():]
	}

	err := vm.InstallModule("arksdk/cli", typescript.Module{
		"args":     strings.Join(extraCLIArgs, " "),
		"argsList": extraCLIArgs,
		"flags":    flags,
	})

	if err != nil {
		return err
	}

	env := make(map[string]string)
	for _, v := range os.Environ() {
		parts := strings.Split(v, "=")
		env[parts[0]] = parts[1]
	}

	return vm.InstallModule("arksdk/os", typescript.Module{
		"env": env,
	})
}

func validateArgsRequired(_ *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("TARGET_PATH is a required parameter")
	}

	return nil
}

type runE func(cmd *cobra.Command, args []string) error

func cobraRunEMiddleware(middleware ...runE) runE {
	return func(cmd *cobra.Command, args []string) error {
		for _, run := range middleware {
			if err := run(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}
