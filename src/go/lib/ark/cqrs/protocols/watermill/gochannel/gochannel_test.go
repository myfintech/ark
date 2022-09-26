package gochannel

import (
	"testing"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/stretchr/testify/require"
)

func TestBroker(t *testing.T) {
	broker := new(Broker)
	require.Implements(t, (*cqrs.Broker)(nil), broker)
}
