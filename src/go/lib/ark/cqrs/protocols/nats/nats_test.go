package nats

import (
	"context"
	"strconv"
	"testing"

	"github.com/myfintech/ark/src/go/lib/utils"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/nats-io/nats.go"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/stretchr/testify/require"
)

func TestBroker(t *testing.T) {
	topic := cqrs.RouteKey("test")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := DefaultServerOptions
	freePort, err := utils.GetFreePort()
	require.NoError(t, err)
	opts.Port, err = strconv.Atoi(freePort)
	require.NoError(t, err)
	instance, err := RunServer(&opts)
	require.NoError(t, err)

	nc, err := nats.Connect(instance.ClientURL())
	require.NoError(t, err)

	broker := NewBroker(nc)
	require.Implements(t, (*cqrs.Broker)(nil), broker)

	message := cqrs.NewDefaultEnvelope(
		cqrs.WithSource("example/uri"),
		cqrs.WithType("test.event"),
		cqrs.WithData(cloudevents.ApplicationJSON, map[string]string{
			"hello": "world",
		}),
	)
	require.NoError(t, message.Error)

	stream, err := broker.Subscribe(ctx, topic, nil)
	require.NoError(t, err)

	err = broker.Publish("test", &message)
	require.NoError(t, err)

	received := <-stream
	require.NoError(t, received.Error)

	data := map[string]string{}
	err = received.DataAs(&data)
	require.NoError(t, err)
	require.NotEmpty(t, data)
}
