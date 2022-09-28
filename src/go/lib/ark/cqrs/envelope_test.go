package cqrs

import (
	"testing"

	"github.com/stretchr/testify/require"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func TestEnvelope(t *testing.T) {
	t.Run("should map data to enveloped message", func(t *testing.T) {
		expectedPayload := map[string]string{
			"hello": "world",
		}

		envelope := NewDefaultEnvelope(
			WithSource("test.sender"),
			WithType("test.event.type"),
			WithData(cloudevents.ApplicationJSON, expectedPayload),
		)
		require.NoError(t, envelope.Error)
		require.NotEmpty(t, envelope.ID())
		require.NotEmpty(t, envelope.Time())
		require.NotEmpty(t, envelope.Data())
		require.Equal(t, envelope.Source(), "test.sender")
		require.Equal(t, envelope.Type(), "test.event.type")

		t.Run("should properly decode message data", func(t *testing.T) {
			actualPayload := map[string]string{}
			require.NoError(t, envelope.DataAs(&actualPayload))
			require.Equal(t, expectedPayload, actualPayload)
		})
	})

	t.Run("should fail when a required fields are not provided", func(t *testing.T) {
		require.Error(t, NewEnvelope().Error)
	})
}
