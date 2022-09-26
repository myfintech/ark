package gochannel

import (
	"context"
	"time"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"

	watermillMessage "github.com/ThreeDotsLabs/watermill/message"
	watermillGoChannel "github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

// Broker implements the cqrs.Broker interface as a wrapper around watermills go channel implementation
type Broker struct {
	wmc *watermillGoChannel.GoChannel
}

// Publish publishes (N) cqrs.Message(s) over the given topic
func (b Broker) Publish(topic cqrs.RouteKey, messages ...cqrs.Message) error {
	for _, message := range messages {
		payload, err := message.MarshalJSON()
		if err != nil {
			return err
		}
		if err = b.wmc.Publish(topic.String(), watermillMessage.NewMessage(message.ID(), payload)); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe attaches to the given topic and returns a channel to receive cqrs.Message(s)
func (b Broker) Subscribe(ctx context.Context, topic cqrs.RouteKey, _ *time.Duration) (<-chan cqrs.Envelope, error) {
	stream := make(chan cqrs.Envelope)
	wmStream, err := b.wmc.Subscribe(ctx, topic.String())
	if err != nil {
		return nil, err
	}

	go func() {
		for wmMessage := range wmStream {
			stream <- cqrs.NewEnvelope(cqrs.FromData(wmMessage.Payload))
		}
	}()
	return stream, nil
}

// Close noop
func (b Broker) Close() error {
	return nil
}

// New creates a new Broker
func New() *Broker {
	return &Broker{
		wmc: watermillGoChannel.NewGoChannel(watermillGoChannel.Config{
			OutputChannelBuffer:            0,
			Persistent:                     false,
			BlockPublishUntilSubscriberAck: false,
		}, nil),
	}
}
