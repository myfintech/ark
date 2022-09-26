package nats

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/nats-io/nats.go"
)

type Broker struct {
	nc *nats.Conn
}

// Publish send a message
func (b Broker) Publish(topic cqrs.RouteKey, messages ...cqrs.Message) error {
	for _, message := range messages {
		data, err := json.Marshal(message)
		if err != nil {
			return err
		}
		err = b.nc.Publish(topic.String(), data)
		if err != nil {
			return err
		}
	}
	return nil
}

// Subscribe registers interest in the given topic and returns a channel to begin processing messages
// The subscription will be canceled and interest will be removed from the topic when the context is canceled
// An error is returned if we fail to subscribe
func (b Broker) Subscribe(
	ctx context.Context,
	topic cqrs.RouteKey,
	_ *time.Duration,
) (<-chan cqrs.Envelope, error) {
	stream := make(chan cqrs.Envelope)
	sub, err := b.nc.Subscribe(topic.String(), func(msg *nats.Msg) {
		stream <- cqrs.NewEnvelope(cqrs.FromData(msg.Data))
	})
	if err != nil {
		close(stream)
		return stream, err
	}

	go func() {
		<-ctx.Done()
		_ = sub.Drain()
	}()

	return stream, nil
}

// Close closes the attaches nats client
func (b Broker) Close() error {
	b.nc.Close()
	return nil
}

// NewBroker creates a new NATS broker
func NewBroker(nc *nats.Conn) *Broker {
	return &Broker{
		nc: nc,
	}
}

func Connect(addr string, options ...nats.Option) (*nats.Conn, error) {
	return nats.Connect(addr, options...)
}

var DefaultServerOptions = server.Options{
	Host:           "127.0.0.1",
	Port:           4222,
	NoLog:          true,
	NoSigs:         true,
	MaxControlLine: 4096,
	HTTPPort:       8222,

	// FIXME: adjust max payload (default 1MB)
	// FB Watchman initial file scan is HUGE
	// MaxPayload:            2e+6,

	// FIXME: enable for message guarantees
	// JetStream: false,

	DisableShortFirstPing: true,
}

func RunServer(opts *server.Options) (*server.Server, error) {
	if opts == nil {
		opts = &DefaultServerOptions
	}

	s, err := server.NewServer(opts)
	if err != nil || s == nil {
		return nil, errors.Wrap(err, "No NATS Server object returned: %v")
	}
	// Run server in Go routine.
	go s.Start()
	// Wait for accept loop(s) to be started
	if !s.ReadyForConnections(10 * time.Second) {
		return nil, errors.New("Unable to start NATS Server in Go Routine")
	}

	return s, nil
}
