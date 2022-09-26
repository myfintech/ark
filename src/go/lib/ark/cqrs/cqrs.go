package cqrs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"
)

// Validator describes an interface for validating a data structure
type Validator interface {
	Validate() error
}

// Message an interface
type Message interface {
	Validator
	json.Marshaler
	json.Unmarshaler
	event.EventReader
	event.EventWriter
}

// Sender is an interface that describes a system that expects to publish (n) Message(s)
type Sender interface {
	Publish(topic RouteKey, messages ...Message) error
}

// Receiver an interface that describes a system that expects to subscribe to a topic and receive Message(s)
type Receiver interface {
	Subscribe(ctx context.Context, topic RouteKey, ackDeadline *time.Duration) (<-chan Envelope, error)
	Close() error
}

// Broker an interface that describes a system that can handle inbound and outbound Message(s)
type Broker interface {
	Sender
	Receiver
}

// Caller is an interface that describes a system that expects to make a synchronous request to another system
type Caller interface {
	Request(topic RouteKey, message Message) (Envelope, error)
}

// OnMessageFunc a function that processes an Envelope message
type OnMessageFunc func(ctx context.Context, msg Envelope) error

// OnMessageErrorFunc a function that processes errors that are returned from OnMessageFunc
type OnMessageErrorFunc func(ctx context.Context, msg Envelope, err error) error
