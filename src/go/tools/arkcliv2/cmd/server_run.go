package cmd

import (
	"github.com/moby/buildkit/util/appcontext"
	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/ark/shared_clients"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems"
	"github.com/myfintech/ark/src/go/lib/ark/workspace"
	"github.com/myfintech/ark/src/go/lib/logz"
	"github.com/reactivex/rxgo/v2"
	"github.com/spf13/cobra"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/protocols/nats"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems/embedded_broker"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems/fs_observer"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems/graph_runner"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems/http_server"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems/k8s_echo"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems/live_sync"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems/port_binder"
	"github.com/myfintech/ark/src/go/lib/fs/observer"
)

func newServerRunCmd(
	serverCmd *cobra.Command,
	fileObserver *observer.Observer,
	subsystemsManager *subsystems.Manager,
	sharedClients *shared_clients.Container,
	store ark.Store,
	logger logz.FieldLogger,
	config *workspace.Config,
) *cobra.Command {
	serverRunCmd := &cobra.Command{
		Use:   "run",
		Short: "ark server run that executes the host server in the foreground",
		RunE: func(cmd *cobra.Command, args []string) error {
			addr, err := cmd.Flags().GetString("address")
			if err != nil {
				return err
			}

			brokerType, err := cmd.Flags().GetString("broker")
			if err != nil {
				return err
			}

			brokerAddress, err := cmd.Flags().GetString("broker-address")
			if err != nil {
				return err
			}

			natsd, err := nats.RunServer(nil)
			if err != nil {
				return err
			}

			disableLiveSync, err := cmd.Flags().GetBool("disable-live-sync")
			if err != nil {
				return err
			}

			conn, err := nats.Connect(natsd.ClientURL())
			if err != nil {
				return err
			}

			fsStream := make(<-chan rxgo.Item)

			if disableLiveSync {
				subsystemsManager.DisabledSubsystems.Store(topics.FSObserver.String(), true)
				subsystemsManager.DisabledSubsystems.Store(
					topics.LiveSyncConnectionManager.String(),
					true,
				)
			} else {
				fsStream = fileObserver.FileSystemStream.Observe()
			}

			broker := nats.NewBroker(conn)
			sharedClients.Broker = broker
			liveSyncConnectionManager := live_sync.NewConnectionManager(appcontext.Context())

			logFilePath, err := logz.SuggestedFilePath("ark", "server.log")
			if err != nil {
				return err
			}

			if err = subsystemsManager.Register(
				http_server.NewSubsystem(addr, logFilePath, store, logger, broker),
				graph_runner.NewSubsystem(store, logger, *sharedClients, broker),
				embedded_broker.NewSubsystem(brokerType, brokerAddress, logger, natsd, broker),
				fs_observer.NewSubsystem(logger, broker, fsStream),
				port_binder.NewSubsystem(broker, logger, sharedClients.K8s),
				port_binder.K8sEchoHandler(broker, logger, sharedClients.K8s),
				live_sync.NewConnectionManagerSubsystem(broker, logger, liveSyncConnectionManager),
				live_sync.NewFSSync(broker, logger, liveSyncConnectionManager, *config),
				k8s_echo.NewSubsystem(broker, logger, sharedClients.K8s, *config),
			); err != nil {
				return err
			}

			if err = subsystemsManager.Start(); err != nil {
				return err
			}

			return subsystemsManager.Wait()
		},
	}
	serverCmd.AddCommand(serverRunCmd)
	serverRunCmd.Flags().
		StringP("address", "a", "127.0.0.1:9000", "The address for the server to listen on")
	serverRunCmd.Flags().
		Bool("disable-live-sync", false, "Disables the live sync system and fs observation (use in CI)")
	return serverRunCmd
}
