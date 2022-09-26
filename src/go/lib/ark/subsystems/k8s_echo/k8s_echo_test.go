package k8s_echo_test

import (
	"context"
	"sync"
	"testing"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/events"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/sources"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"
	"github.com/myfintech/ark/src/go/lib/ark/shared_clients"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems/k8s_echo"
	"github.com/myfintech/ark/src/go/lib/ark/workspace"
	"github.com/myfintech/ark/src/go/lib/logz"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestNewSubsystem(t *testing.T) {
	broker := cqrs.NewMockBroker()
	logger := new(logz.MockLogger)
	broker.On("Subscribe", topics.K8sEchoEvents)
	broker.On("Publish", topics.K8sEchoEvents)
	logger.On("Child", mock.Anything)
	logger.On("Info", mock.Anything)
	logger.On("Debugf", mock.Anything, mock.Anything)
	client, err := shared_clients.NewContainerWithDefaults()
	require.NoError(t, err)

	ctx, shutdown := context.WithCancel(context.Background())
	wg := new(sync.WaitGroup)
	eg, egCTX := errgroup.WithContext(ctx)

	inbox, err := broker.Subscribe(ctx, topics.K8sEchoEvents, nil)
	require.NoError(t, err)
	config := new(workspace.Config)
	wg.Add(1)
	eg.Go(
		k8s_echo.NewSubsystem(broker, logger, client.K8s, *config).Factory(wg, egCTX),
	)
	wg.Wait()

	// publish a message that the subsystem should react to
	err = broker.Publish(
		topics.K8sEchoEvents,
		cqrs.NewDefaultEnvelope(
			sources.K8sEcho,
			events.K8sEchoPodChangedType,
			cqrs.WithData(cqrs.TextPlain, "test")),
	)
	require.NoError(t, err)

	// wait for the subsystem to publish an error message
	envelope := <-inbox
	require.NotEmpty(t, envelope.Data())
	require.Equal(t, envelope.Type(), events.K8sEchoPodChanged.String())

	// shutdown the subsystem
	shutdown()

	broker.AssertExpectations(t)
	require.NoError(t, eg.Wait())
}
