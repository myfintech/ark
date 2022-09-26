package cqrs

import (
	"context"
	"time"
)

type NoOpBroker struct{}

func (m NoOpBroker) Publish(_ RouteKey, _ ...Message) error {
	return nil
}

func (m NoOpBroker) Subscribe(_ context.Context, _ RouteKey, _ *time.Duration) (<-chan Envelope, error) {
	stream := make(chan Envelope)
	close(stream)
	return stream, nil
}

func (m NoOpBroker) Close() error {
	return nil
}
