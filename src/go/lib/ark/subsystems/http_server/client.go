package http_server

import (
	"io"

	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/messages"
	"github.com/myfintech/ark/src/go/lib/dag"
)

// Client represent a struct
type Client interface {
	AddTarget(target ark.RawTarget) (ark.RawArtifact, error)
	GetTargets() ([]ark.RawTarget, error)
	ConnectTargets(edge ark.GraphEdge) (ark.GraphEdge, error)
	GetGraph() (*dag.AcyclicGraph, error)
	GetGraphEdges() ([]ark.GraphEdge, error)
	Run(cmd messages.GraphRunnerExecuteCommand) (messages.GraphRunnerExecuteCommandResponse, error)
	GetServerLogs() (io.Reader, error)
	GetLogsByKey(logKey string) (io.Reader, error)
}
