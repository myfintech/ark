package cmd

import (
	"bytes"
	"context"
	"net"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/myfintech/ark/src/go/lib/fs"

	"github.com/myfintech/ark/src/go/lib/fs/observer"

	"google.golang.org/api/option"
	"google.golang.org/api/transport/grpc"

	"github.com/myfintech/ark/src/go/lib/ark/components/entrypoint"

	"github.com/moby/buildkit/util/appcontext"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/deploy"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/myfintech/ark/src/go/lib/log"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
)

func getValidTargets(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	if err := decodeWorkspacePreRunE(cmd, args); err != nil {
		log.WithError(err).Error("Failed to load workspace")
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var completions []string

addressLoop:
	for _, address := range workspace.TargetLUT.SortedAddresses() {
		for _, arg := range args {
			if arg == address {
				continue addressLoop
			}
		}
		if strings.HasPrefix(address, toComplete) {
			completions = append(completions, address)
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

func registerObservableTarget(ctx context.Context, rootTarget base.Addressable, grpcClient entrypoint.SyncClient, stopOnFirstError bool) error {
	streamChangeClient, streamErr := grpcClient.StreamFileChange(ctx)
	if streamErr != nil {
		return streamErr
	}
	log.Info("streaming file changes")

	observable := workspace.WatchForChanges(base.FilterChangeNotificationsByTarget(
		workspace.TargetGraph.Isolate(rootTarget), rootTarget),
		func(ctx context.Context, i interface{}) (interface{}, error) {
			notification := i.(*observer.ChangeNotification)
			var files []*entrypoint.File
			var actions []*entrypoint.Action
			var filePaths []string
			for _, file := range notification.Files {
				files = append(files, &entrypoint.File{
					Name:          file.Name,
					Exists:        file.Exists,
					New:           file.New,
					Type:          file.Type,
					Hash:          file.Hash,
					SymlinkTarget: file.SymlinkTarget,
					RelName:       file.RelName,
				})
				filePaths = append(filePaths, file.Name)
				log.Infof("%s %s", file.Name, file.Type)

				buildActions, err := rootTarget.(deploy.Target).ActionsToSend(file.Name)
				if err != nil {
					return i, errors.Wrap(err, "unable to get actions to send with change notification")
				}
				actions = append(actions, buildActions...)
			}

			archiveBuffer := bytes.NewBuffer([]byte{})

			gzipErr := fs.GzipTarFiles(filePaths, workspace.Dir, archiveBuffer, nil)
			if gzipErr != nil {
				return i, gzipErr
			}

			sendErr := streamChangeClient.Send(&entrypoint.FileChangeNotification{
				Files:   files,
				Archive: archiveBuffer.Bytes(),
				Actions: actions,
			})
			if sendErr != nil {
				return i, errors.Wrap(sendErr, "send error")
			}

			_, recErr := streamChangeClient.Recv()
			return i, errors.Wrap(recErr, "receive error")
		})
	for message := range observable.Observe() {
		if message.Error() {
			log.Errorf("change stream error %s", message.E)
			if stopOnFirstError {
				return message.E
			}
		}
	}
	return nil
}

var runCmd = &cobra.Command{
	Use:               "run TARGET_ADDRESS",
	Short:             "run is a sub-command of ark that attempts to achieve target state",
	PersistentPreRunE: decodeWorkspacePreRunE,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetAddress := args[0]

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		push, err := cmd.Flags().GetBool("push")
		if err != nil {
			return err
		}

		watch, err := cmd.Flags().GetBool("watch")
		if err != nil {
			return err
		}
		workspace.TargetWatch = watch

		stopOnFirstError, err := cmd.Flags().GetBool("stop_on_first_error")
		if err != nil {
			return err
		}

		rootTarget, err := workspace.TargetLUT.LookupByAddress(targetAddress)
		if err != nil {
			return errors.Wrap(err, "cannot lookup root target by the address given")
		}

		push = push && workspace.Config.Artifacts != nil
		pull := workspace.Config.Artifacts != nil
		if err = workspace.GraphWalk(rootTarget.Address(), base.BuildWalker(force, pull, push)); err != nil {
			return err
		}

		if !watch {
			return nil
		}

		eg, ctx := errgroup.WithContext(appcontext.Context())

		for command := range workspace.ReadyPortCommands {
			eg.Go(func() error {
				if binding, ok := command.PortMap["ark_grpc_entrypoint"]; ok {
					client, clientErr := getGRPCClient(ctx, binding.HostPort)
					if clientErr != nil {
						return clientErr
					}

					if targetErr := registerObservableTarget(ctx, rootTarget, client, stopOnFirstError); targetErr != nil {
						return targetErr
					}
					return nil
				}
				return nil
			})
		}

		return eg.Wait()
	},
	ValidArgsFunction: getValidTargets,
}

func init() {
	rootCmd.AddCommand(runCmd)
	_ = viper.BindPFlags(rootCmd.PersistentFlags())
	runCmd.Flags().Bool("force", false, "Forces a build regardless of cached state")
	runCmd.Flags().Bool("watch", false, "Watches the file system and re-builds targets based on file changes")
	runCmd.Flags().Bool("push", false, "Attempts to push artifacts if a remote cache is configured for the workspace")
	runCmd.Flags().Bool("stop_on_first_error", false, "While in watch mode, stop on the first error")
}

func getGRPCClient(ctx context.Context, grpcPort string) (entrypoint.SyncClient, error) {
	conn, connErr := grpc.DialInsecure(ctx, option.WithEndpoint(net.JoinHostPort("localhost", grpcPort)))
	if connErr != nil {
		return nil, connErr
	}
	return entrypoint.NewSyncClient(conn), nil
}

// move context into run var and pass it to get grpc func; update get grpc func to take in ctx
