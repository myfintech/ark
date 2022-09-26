package graph

import (
	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/dag"
	"github.com/pkg/errors"
)

// FromStore accepts an ark store and produces a DAG from targets and edges
func FromStore(store ark.Store) (*dag.AcyclicGraph, error) {
	graph := new(dag.AcyclicGraph)
	targetsCache := make(map[string]ark.RawTarget)

	targets, err := store.GetTargets()
	if err != nil {
		return graph, err
	}

	for _, target := range targets {
		graph.Add(target)
		targetsCache[target.Key()] = target
	}

	edges, err := store.GetGraphEdges()
	if err != nil {
		return graph, err
	}

	for _, edge := range edges {
		src, exists := targetsCache[edge.Src]
		if !exists {
			return graph, errors.Errorf("src edge with key %s does not exist", edge.Src)
		}
		dst, exists := targetsCache[edge.Dst]
		if !exists {
			return graph, errors.Errorf("dst edge with key %s does not exist", edge.Dst)
		}

		graph.Connect(dag.BasicEdge(src, dst))
	}

	return graph, nil
}
