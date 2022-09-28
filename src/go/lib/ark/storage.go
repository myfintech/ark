package ark

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/myfintech/ark/src/go/lib/dag"
)

// Store ...
type Store interface {
	GetTargets() ([]RawTarget, error)
	GetTargetByKey(key string) (RawTarget, error)
	AddTarget(target RawTarget) (RawArtifact, error)
	ConnectTargets(edge GraphEdge) error
	GetGraph() (*dag.AcyclicGraph, error)
	GetGraphEdges() ([]GraphEdge, error)
	Open(connection string) error
	Migrate() error
}

// GraphEdge represents the src key and dst key between two targets in a graph
type GraphEdge struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
}

// Validate ensure graph edge is valid
func (g *GraphEdge) Validate() error {
	return validation.ValidateStruct(g,
		validation.Field(&g.Src, validation.Required),
		validation.Field(&g.Dst, validation.Required),
	)
}

// Key returns a key that can be used to index edges
func (g GraphEdge) Key() string {
	return fmt.Sprintf("%s:%s", g.Src, g.Dst)
}
