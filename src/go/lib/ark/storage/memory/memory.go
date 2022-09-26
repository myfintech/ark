package memory

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/ark/derivation"

	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/ark/storage/graph"
	"github.com/myfintech/ark/src/go/lib/dag"
)

// Store an in memory implementation of the ark.Store
type Store struct {
	targets    sync.Map
	graphEdges sync.Map
}

// GetTargets returns a list of []ark.RawTarget from the memory state
func (s *Store) GetTargets() (targets []ark.RawTarget, err error) {
	s.targets.Range(func(key, t interface{}) bool {
		targets = append(targets, t.(ark.RawTarget))
		return true
	})
	return
}

// GetTargetByKey returns a target by its key with an error if it doesn't exist or if we fail to cast a target to the correct type
func (s *Store) GetTargetByKey(key string) (target ark.RawTarget, err error) {
	if v, ok := s.targets.Load(key); ok {
		if target, ok = v.(ark.RawTarget); !ok {
			err = errors.Errorf("target with key %s was not type ark.Target it was %T", key, v)
			return
		}
		return
	}

	err = errors.Errorf("failed to locate target by key %s", key)
	return
}

// AddTarget adds a target to to the in memory storage
func (s *Store) AddTarget(target ark.RawTarget) (artifact ark.RawArtifact, err error) {
	if err = target.Validate(); err != nil {
		return
	}
	s.targets.Store(target.Key(), target)
	for _, ancestor := range target.DependsOn {
		err = s.ConnectTargets(ark.GraphEdge{
			Src: target.Key(),
			Dst: ancestor.Key,
		})
		if err != nil {
			return
		}
	}
	return derivation.RawArtifactFromRawTarget(target)
}

// ConnectTargets adds a graph edge to memory state
func (s *Store) ConnectTargets(edge ark.GraphEdge) error {
	if err := edge.Validate(); err != nil {
		return err
	}
	s.graphEdges.Store(edge.Key(), edge)
	return nil
}

// GetGraphEdges returns the list of graph edges in memory
func (s *Store) GetGraphEdges() (edges []ark.GraphEdge, err error) {
	s.graphEdges.Range(func(key, e interface{}) bool {
		edges = append(edges, e.(ark.GraphEdge))
		return true
	})
	return
}

// GetGraph calculates a DAG from the set of targets and edges
func (s *Store) GetGraph() (*dag.AcyclicGraph, error) {
	return graph.FromStore(s)
}

// Open is a noop in the memory Store
func (s *Store) Open(_ string) error {
	return nil
}

// Migrate is a noop in the memory Store
func (s *Store) Migrate() error {
	return nil
}

var loadStoreOnce sync.Once
var cachedMemoryStore *Store

// NewCachedMemoryStore returns a singleton memory store (idempotent)
func NewCachedMemoryStore() *Store {
	loadStoreOnce.Do(func() {
		cachedMemoryStore = new(Store)
	})
	return cachedMemoryStore
}
