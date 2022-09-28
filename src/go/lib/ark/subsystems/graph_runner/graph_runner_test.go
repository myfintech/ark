package graph_runner

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/events"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"

	"github.com/myfintech/ark/src/go/lib/ark/shared_clients"
	"github.com/myfintech/ark/src/go/lib/ark/storage/memory"
	"github.com/myfintech/ark/src/go/lib/logz"

	"golang.org/x/sync/errgroup"

	"github.com/stretchr/testify/require"
)

func TestNewSubsystem(t *testing.T) {
	ctx, shutdown := context.WithCancel(context.Background())
	broker := cqrs.NewMockBroker()

	shareClients, err := shared_clients.NewContainerWithDefaults()
	require.NoError(t, err)

	store := new(memory.Store)
	logger := new(logz.MockLogger)

	wg := new(sync.WaitGroup)
	eg, egCTX := errgroup.WithContext(ctx)

	// the subsystem should subscribe to the correct command topic
	broker.On("Subscribe", topics.GraphRunnerEvents)
	broker.On("Subscribe", topics.GraphRunnerCommands)
	broker.On("Publish", topics.GraphRunnerCommands)
	broker.On("Publish", topics.GraphRunnerEvents)
	logger.On("Debugf", mock.Anything, mock.Anything)

	inbox, err := broker.Subscribe(ctx, topics.GraphRunnerEvents, nil)
	require.NoError(t, err)

	wg.Add(1)
	eg.Go(NewSubsystem(store, logger, *shareClients, broker).Factory(wg, egCTX))
	wg.Wait()

	// publish a message that the subsystem should react to
	err = broker.Publish(
		topics.GraphRunnerCommands,
		cqrs.NewDefaultEnvelope(
			cqrs.WithData(cqrs.TextPlain, "hello world")),
	)
	require.NoError(t, err)

	// wait for the subsystem to publish an error message
	envelope := <-inbox
	require.NotEmpty(t, envelope.Data())
	require.Equal(t, envelope.Type(), events.GraphRunnerFailed.String())

	// shutdown the subsystem
	shutdown()

	broker.AssertExpectations(t)
	require.NoError(t, eg.Wait())

}
