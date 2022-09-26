package cqrs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRouteKey(t *testing.T) {
	t.Run("should join a single key", func(t *testing.T) {
		system := RouteKey("test.runner")
		commands := system.With("commands")
		require.Equal(t, commands.String(), "test.runner.commands")
	})

	t.Run("should join multiple keys", func(t *testing.T) {
		key := RouteKey("test.runner").With("events").With("example", "event", "worked")
		require.Equal(t, key.String(), "test.runner.events.example.event.worked")
	})

}
