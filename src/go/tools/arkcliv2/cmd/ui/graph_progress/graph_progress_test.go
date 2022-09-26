package graph_progress

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/events"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"

	"github.com/charmbracelet/bubbles/spinner"
)

func TestTargetModel(t *testing.T) {
	model := &TargetModel{
		state:     queued,
		name:      "src/go/tools/test/build.ts:image",
		hash:      hex.EncodeToString(sha256.New().Sum(nil)),
		lastEvent: events.GraphWalkerStarted,
		spinner: spinner.Model{
			Spinner: spinner.Dot,
		},
	}

	require.Implements(t, (*tea.Model)(nil), model)
	t.Log(model.View())
}
