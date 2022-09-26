package cqrs

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockBroker struct {
	topics map[RouteKey][]chan Envelope
	mock.Mock
}

func (m *MockBroker) Publish(topic RouteKey, messages ...Message) error {
	m.Called(topic)
	for _, inbox := range m.topics[topic] {
		for _, message := range messages {
			inbox <- message.(Envelope)
		}
	}
	return nil
}

func (m *MockBroker) Subscribe(
	_ context.Context,
	topic RouteKey,
	_ *time.Duration,
) (<-chan Envelope, error) {
	m.Called(topic)
	if _, ok := m.topics[topic]; !ok {
		subStream := make(chan Envelope, 1000)
		m.topics[topic] = append(m.topics[topic], subStream)
	}
	return m.topics[topic][len(m.topics[topic])-1], nil
}

func (m *MockBroker) Close() error {
	m.Called()
	return nil
}

func NewMockBroker() *MockBroker {
	broker := &MockBroker{
		topics: make(map[RouteKey][]chan Envelope),
	}
	return broker
}
