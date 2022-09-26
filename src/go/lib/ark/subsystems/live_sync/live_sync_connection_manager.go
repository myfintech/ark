package live_sync

import (
	"context"
	"net"
	"sync"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/events"
	"github.com/myfintech/ark/src/go/lib/kube/portbinder"

	"github.com/pkg/errors"

	"google.golang.org/grpc"

	"github.com/myfintech/ark/src/go/lib/ark/components/entrypoint"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems"
	"github.com/myfintech/ark/src/go/lib/logz"
)

// ConnectionManager is the client struct for the live sync subsystem
type ConnectionManager struct {
	ctx         context.Context
	connections sync.Map
}

// Len returns the count of connections
func (c *ConnectionManager) Len() (count int) {
	c.connections.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return
}

// Connect uses gRPC to dial the given address and stores the client in the connections sync map
func (c *ConnectionManager) Connect(key, address string) error {
	client, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return err
	}

	syncClient := entrypoint.NewSyncClient(client)
	streamClient, err := syncClient.StreamFileChange(c.ctx)
	if err != nil {
		return err
	}

	c.connections.Store(key, streamClient)

	return nil
}

// Disconnect closes the client connection
func (c *ConnectionManager) Disconnect(key string) error {
	connection, err := c.loadClient(key)
	if err != nil {
		return nil
	}

	return connection.CloseSend()
}

// GetAllConnections returns a list of ready connections to connect with
func (c *ConnectionManager) GetAllConnections() (connections []entrypoint.Sync_StreamFileChangeClient) {
	c.connections.Range(func(key, value interface{}) bool {
		connections = append(connections, value.(entrypoint.Sync_StreamFileChangeClient))
		return true
	})
	return
}

func (c *ConnectionManager) loadClient(key string) (entrypoint.Sync_StreamFileChangeClient, error) {
	loadedClient, ok := c.connections.Load(key)
	if !ok {
		return nil, errors.Errorf("there was an error loading the client from the sync map: %s", key)
	}

	return loadedClient.(entrypoint.Sync_StreamFileChangeClient), nil
}

// NewConnectionManager creates a pointer to a connection manager with its parent context set
func NewConnectionManager(ctx context.Context) *ConnectionManager {
	return &ConnectionManager{
		ctx: ctx,
	}
}

// NewConnectionManagerSubsystem waits for port binder events and creates streaming GRPC connections for the live sync subsystem
func NewConnectionManagerSubsystem(broker cqrs.Broker, logger logz.FieldLogger, manager *ConnectionManager) *subsystems.Process {
	logger = logger.Child(logz.WithFields(logz.Fields{
		"system": "live.sync.connection.manager",
	}))
	return &subsystems.Process{
		Name:    topics.LiveSyncConnectionManager.String(),
		Factory: subsystems.Reactor(topics.PortBinderEvents, broker, logger, newConnectionManagerMessageHandler(manager, broker, logger), newConnectionManagerErrorHandler(), nil),
	}
}

func newConnectionManagerMessageHandler(manager *ConnectionManager, _ cqrs.Broker, logger logz.FieldLogger) cqrs.OnMessageFunc {
	logger.Info("ready")
	return func(ctx context.Context, msg cqrs.Envelope) error {
		if msg.Type() != events.PortBinderSuccess.String() {
			return nil
		}

		if msg.Error != nil {
			return msg.Error
		}

		var cmd portbinder.BindPortCommand

		if err := msg.DataAs(&cmd); err != nil {
			return err
		}

		binding, ok := cmd.PortMap["ark_grpc_entrypoint"]
		if !ok {
			return nil
		}

		addr := net.JoinHostPort("localhost", binding.HostPort)

		if err := manager.Connect(addr, addr); err != nil {
			return err
		}

		return nil
	}
}

func newConnectionManagerErrorHandler() func(ctx context.Context, msg cqrs.Envelope, err error) error {
	return func(ctx context.Context, msg cqrs.Envelope, err error) error {
		return err
	}
}
