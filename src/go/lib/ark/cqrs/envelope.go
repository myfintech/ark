package cqrs

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	ceSDK "github.com/cloudevents/sdk-go/v2"
	ceEvent "github.com/cloudevents/sdk-go/v2/event"
)

const (
	TextPlain       = ceEvent.TextPlain
	ApplicationJSON = ceEvent.ApplicationJSON
)

// Envelope a value object (container) for Message
// Allows a recipient to check if a message contained an error
type Envelope struct {
	Message
	Error  error
	ack    chan error
	Header http.Header
}

// Ack closes the ack channel without an error
func (e *Envelope) Ack() {
	close(e.ack)
}

// Reject rejects the message with an error
func (e *Envelope) Reject(err error) {
	select {
	case e.ack <- err:
		close(e.ack)
	default:
	}
}

// Wait waits until the ack channel has been closed
func (e *Envelope) Wait() <-chan error {
	return e.ack
}

// TypeKey returns the RouteKey of the event type
func (e *Envelope) TypeKey() RouteKey {
	return RouteKey(e.Type())
}

// SourceKey returns the RouteKey of the event source
func (e *Envelope) SourceKey() RouteKey {
	return RouteKey(e.Source())
}

// SubjectKey returns the RouteKey of the event subject
func (e *Envelope) SubjectKey() RouteKey {
	return RouteKey(e.Subject())
}

// NewMessage creates a new instance compatible Message
func NewMessage() Message {
	message := ceSDK.NewEvent()
	return &message
}

// EnvelopeOption a higher order function that modifies properties of a message
type EnvelopeOption func(envelope *Envelope) error

// WithID sets the ID property of the enveloped Message
func WithID(messageID string) EnvelopeOption {
	return func(envelope *Envelope) error {
		envelope.SetID(messageID)
		return nil
	}
}

// WithNewUUID sets the ID property of the enveloped Message to a random UUIDv4
func WithNewUUID() EnvelopeOption {
	return WithID(uuid.New().String())
}

// WithType sets the type property of the enveloped Message
func WithType(messageType RouteKey) EnvelopeOption {
	return func(envelope *Envelope) error {
		envelope.SetType(messageType.String())
		return nil
	}
}

// WithTime sets the time of the enveloped Message
func WithTime(ts time.Time) EnvelopeOption {
	return func(envelope *Envelope) error {
		envelope.SetTime(ts)
		return nil
	}
}

// WithSubject sets the subject of the enveloped Message
func WithSubject(subject RouteKey) EnvelopeOption {
	return func(envelope *Envelope) error {
		envelope.SetSubject(subject.String())
		return nil
	}
}

// WithSource sets the source of the enveloped Message
func WithSource(source RouteKey) EnvelopeOption {
	return func(envelope *Envelope) error {
		envelope.SetSource(source.String())
		return nil
	}
}

// WithData encodes the given payload with the given content type.
// If the provided payload is a byte array, when marshalled to json it will be encoded as base64.
// If the provided payload is different from byte array, datacodec.Encode is invoked to attempt a
// marshalling to byte array.
func WithData(encoding string, data interface{}) EnvelopeOption {
	return func(envelope *Envelope) error {
		return envelope.SetData(encoding, data)
	}
}

func WithHeaders(header http.Header) EnvelopeOption {
	return func(envelope *Envelope) error {
		envelope.Header = header
		return nil
	}
}

// FromData accepts raw byte data an attempts to json.Unmarshal into the enveloped Message
// If an error occurs while unmarshalling it will be accessible from Envelope.Error
func FromData(data []byte) EnvelopeOption {
	return func(envelope *Envelope) error {
		return json.Unmarshal(data, &envelope.Message)
	}
}

// NewEnvelope accepts raw message bytes, attempts to unmarshall the event into a Message
// If there is a problem decoding the message the error will be stored on the Envelope
func NewEnvelope(options ...EnvelopeOption) (envelope Envelope) {
	envelope.ack = make(chan error, 1)
	envelope.Message = NewMessage()
	envelope.Header = http.Header{}

	for _, option := range options {
		if envelope.Error = option(&envelope); envelope.Error != nil {
			return
		}
	}

	envelope.Error = envelope.Validate()

	return
}

// NewDefaultEnvelope accepts raw message bytes, attempts to unmarshall the event into a Message
// If there is a problem decoding the message the error will be stored on the Envelope
// Sets message.ID and Time
func NewDefaultEnvelope(options ...EnvelopeOption) (envelope Envelope) {
	options = append([]EnvelopeOption{
		WithNewUUID(),
		WithTime(time.Now()),
		WithHeaders(http.Header{}),
	}, options...)
	return NewEnvelope(
		options...,
	)
}
