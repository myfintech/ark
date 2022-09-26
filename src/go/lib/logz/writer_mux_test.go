package logz

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/require"
)

type mockTransport struct {
	name   string
	stream chan ThreadSaveEntry
	mock.Mock
}

func (m *mockTransport) Stream() chan ThreadSaveEntry {
	return m.stream
}

func (m *mockTransport) Name() string {
	return m.name
}

func (m *mockTransport) Write(e Entry) error {
	m.Called(e.Message, e.Level)
	return nil
}

func (m *mockTransport) Cleanup() error {
	m.Called()
	return nil
}

func newBuilder(t *mockTransport) Builder {
	return func() (transport Transport, err error) {
		return t, nil
	}
}

func TestMux(t *testing.T) {
	t.Run("should guarantee delivery of all log messages to a single observer", func(t *testing.T) {
		maxCalls := 100
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		logger := New(ctx)

		testCases := []struct {
			transport *mockTransport
			maxCalls  int
		}{
			{
				transport: &mockTransport{name: "t1", stream: make(chan ThreadSaveEntry, 1000)},
			},
			{
				transport: &mockTransport{name: "t2", stream: make(chan ThreadSaveEntry, 1000)},
			},
		}

		for _, testCase := range testCases {
			err := WithMux(newBuilder(testCase.transport))(logger)
			require.NoError(t, err)
			testCase.transport.On("Cleanup").Return(nil)
			testCase.transport.On("Write", mock.Anything, mock.Anything).Return(nil).Times(maxCalls)
		}

		for i := 0; i < maxCalls; i++ {
			logger.Info("testing", i)
		}

		logger.Close()
		require.NoError(t, logger.Wait())

		for _, testCase := range testCases {
			testCase.transport.AssertExpectations(t)
		}
	})

	t.Run("should not return an error when transport context is canceled or times out", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		logger := New(ctx)

		time.AfterFunc(time.Second*2, func() {
			cancel()
			logger.Info("testing")
		})

		err := logger.Wait()
		require.NoError(t, err)
	})

	t.Run("should not return an error when logger is closed naturally", func(t *testing.T) {
		ctx := context.Background()
		logger := New(ctx)

		logger.Close()
		err := logger.Wait()
		require.NoError(t, err)
	})
}
