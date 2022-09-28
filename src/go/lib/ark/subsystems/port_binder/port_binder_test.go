package port_binder

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/events"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"
	"github.com/myfintech/ark/src/go/lib/ark/shared_clients"
	"github.com/myfintech/ark/src/go/lib/logz"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestPortBinderSubsystem(t *testing.T) {
	ctx, shutdown := context.WithCancel(context.Background())
	broker := cqrs.NewMockBroker()

	sharedClients, err := shared_clients.NewContainerWithDefaults()
	require.NoError(t, err)

	logger := new(logz.MockLogger)

	wg := new(sync.WaitGroup)
	eg, egCTX := errgroup.WithContext(ctx)

	// the subsystem should subscribe to the correct command topic
	broker.On("Subscribe", topics.PortBinderCommands)
	broker.On("Subscribe", topics.PortBinderEvents)
	broker.On("Subscribe", topics.K8sEchoEvents)
	broker.On("Publish", topics.PortBinderCommands)
	broker.On("Publish", topics.PortBinderEvents)
	broker.On("Publish", topics.K8sEchoEvents)
	logger.On("Child", mock.Anything)
	logger.On("Info", mock.Anything)
	logger.On("Error", mock.Anything)
	logger.On("Debugf", mock.Anything, mock.Anything)

	portBinderInbox, err := broker.Subscribe(ctx, topics.PortBinderEvents, nil)
	require.NoError(t, err)
	k8sEchoInbox, errK8sEchoInbox := broker.Subscribe(ctx, topics.K8sEchoEvents, nil)
	require.NoError(t, errK8sEchoInbox)

	wg.Add(1)
	eg.Go(NewSubsystem(broker, logger, sharedClients.K8s).Factory(wg, egCTX))
	// eg.Go(K8sEchoHandler(broker, logger, sharedClients.K8s).Factory(wg, egCTX))
	wg.Wait()

	// publish a message that the subsystem should react to
	err = broker.Publish(
		topics.PortBinderCommands,
		cqrs.NewDefaultEnvelope(cqrs.WithData(cqrs.TextPlain, "hello world")),
	)
	require.NoError(t, err)

	// wait for the subsystem to publish an error message
	envelopePortBinder := <-portBinderInbox
	require.NotEmpty(t, envelopePortBinder.Data())
	require.Equal(t, envelopePortBinder.Type(), events.PortBinderFailed.String())

	// // publish a message that the subsystem should react to
	errK8sEchoInbox = broker.Publish(
		topics.K8sEchoEvents,
		cqrs.NewDefaultEnvelope(cqrs.WithData(cqrs.TextPlain, "hello world")),
	)
	require.NoError(t, errK8sEchoInbox)

	// // wait for the subsystem to publish an error message
	envelopeK8sEcho := <-k8sEchoInbox
	require.NotEmpty(t, envelopeK8sEcho.Data())
	require.Equal(t, envelopePortBinder.Type(), events.PortBinderFailed.String())

	// shutdown the subsystem
	shutdown()

	broker.AssertExpectations(t)
	require.NoError(t, eg.Wait())
}
