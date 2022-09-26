package base

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform/tfdiags"
	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/dag"
	"github.com/myfintech/ark/src/go/lib/log"
)

// GraphWalker a callback function used to process targets in the order they should be executed in based on their dependency chain
type GraphWalker func(dag.Vertex) error

// Graph a directed acyclic graph of build targets
type Graph struct {
	dag dag.AcyclicGraph
}

// DirectedGraph returns the internal dag
func (g *Graph) DirectedGraph() dag.AcyclicGraph {
	return g.dag
}

// Add adds a target to the graph
func (g *Graph) Add(vertex dag.Vertex) dag.Vertex {
	return g.dag.Add(vertex)
}

// Connect creates an edge between a source and destination target
func (g *Graph) Connect(src, dest dag.Vertex) {
	g.dag.Connect(dag.BasicEdge(src, dest))
}

// Roots returns a all roots in the graph
func (g *Graph) Roots() []dag.Vertex {
	roots := make([]dag.Vertex, 0)
	for _, v := range g.dag.Vertices() {
		if g.dag.UpEdges(v).Len() == 0 {
			roots = append(roots, v)
		}
	}
	return roots
}

// Validate ensures the graph is an acyclic graph (no cycles)
func (g *Graph) Validate() error {
	var err error
	cycles := g.dag.Cycles()
	if len(cycles) > 0 {
		for _, cycle := range cycles {
			cycleStr := make([]string, len(cycle))
			for j, vertex := range cycle {
				cycleStr[j] = dag.VertexName(vertex)
			}

			err = multierror.Append(err, fmt.Errorf(
				"cycle: %s", strings.Join(cycleStr, ", ")))
		}
	}

	// Look for cycles to self
	for _, e := range g.dag.Edges() {
		if e.Source() == e.Target() {
			err = multierror.Append(err, fmt.Errorf(
				"self reference: %s", dag.VertexName(e.Source())))
		}
	}

	return err
}

// TopologicalSort of a directed graph is a linear ordering of its vertices such that
// for every directed edge uv from vertex u to vertex v, u comes before v in the ordering
// This function sorts the vertices by their name to be deterministic.
func (g *Graph) TopologicalSort(start dag.Vertex) []dag.Vertex {
	return g.dag.TopologicalSort(start)
}

// Isolate returns an isolated sub graph from the starting vertex
func (g *Graph) Isolate(start dag.Vertex) *Graph {
	graph := new(Graph)
	vertices := make([]dag.Vertex, 0)
	_ = g.dag.DepthFirstWalk([]dag.Vertex{start}, func(vertex dag.Vertex, i int) error {
		vertices = append(vertices, vertex)
		return nil
	})
	for _, vertex := range vertices {
		graph.Add(vertex)
		for _, edge := range g.dag.EdgesFrom(vertex) {
			graph.Add(edge.Source())
			graph.Add(edge.Target())
			graph.Connect(edge.Source(), edge.Target())
		}
	}
	graph.dag.TransitiveReduction()
	return graph
}

// TransitiveReduction performs the transitive reduction of graph g in place.
func (g *Graph) TransitiveReduction() {
	g.dag.TransitiveReduction()
}

// Walk walks the graph, calling your callback as each node is visited.
// This will walk nodes in parallel if it can. The resulting diagnostics
// contains problems from all graphs visited, in no particular order.
func (g *Graph) Walk(walk GraphWalker) error {
	return g.dag.Walk(func(vertex dag.Vertex) tfdiags.Diagnostics {
		diag := new(tfdiags.Diagnostics)
		if err := walk(vertex); err != nil {
			return diag.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  err.Error(),
				Detail:   err.Error(),
			})
		}
		return nil
	}).Err()
}

// BuildWalker walks the graph and builds each target sequentially attempting to cache where possible
func BuildWalker(force bool, pullFromRemoteCache bool, pushToRemoteCache bool) GraphWalker {
	return func(vertex dag.Vertex) error {
		buildableTarget, buildable := vertex.(Buildable)
		cacheableTarget, cacheable := vertex.(Cacheable)
		if !buildable {
			return errors.Errorf("%v(%s) of type %T does not implement the Buildable interface", vertex, vertex, vertex)
		}
		// cacheable = cacheable && cacheableTarget.CacheEnabled()

		address := buildableTarget.Address()
		shortHash := buildableTarget.ShortHash()
		ctxLogger := log.WithFields(log.Fields{
			"prefix": fmt.Sprintf("%s %s", shortHash, address),
		})

		pushIfApplicable := func() error {
			if cacheable && pushToRemoteCache {
				// TODO: should we be pushing if the artifact is already remotely cached?
				// --force-push
				remotelyCached, err := cacheableTarget.CheckRemoteCache()
				if err != nil {
					return errors.Wrapf(err, "failed to verify remote cache %s", address)
				}
				if remotelyCached {
					return nil
				}
				ctxLogger.Info("pushing artifacts")
				if err = cacheableTarget.PushRemoteCache(); err != nil {
					return errors.Wrapf(err, "failed to push %s to remote cache", address)
				}
			}
			return nil
		}

		build := func() error {
			ctxLogger.Info("building artifacts")
			if err := buildableTarget.Build(); err != nil {
				return errors.Wrapf(err, "failed to build %s", address)
			}

			if cacheable {
				ctxLogger.Info("saving local build cache state")
				if err := cacheableTarget.SaveLocalBuildCacheState(); err != nil {
					return errors.Wrapf(err, "failed to save local cache state %s", address)
				}
			}

			return pushIfApplicable()
		}

		if err := buildableTarget.PreBuild(); err != nil {
			return errors.Wrapf(err, "failed prebuild %s", address)
		}

		if !cacheable || force {
			ctxLogger.Info("force building")
			return build()
		}

		ctxLogger.Info("verifying local build cache")
		locallyCached, err := cacheableTarget.CheckLocalBuildCache()
		if err != nil {
			return errors.Wrapf(err, "failed to verify local build cache for %s", address)
		}

		switch {
		case locallyCached:
			ctxLogger.Info("target was locally cached, skipping build")
			return pushIfApplicable()
		case (!locallyCached && !pullFromRemoteCache) || !pullFromRemoteCache: // this particular check is necessary because a remote artifact store might not be configured
			ctxLogger.Info("target was not cached or pulling is disabled, executing build")
			return build()
		}

		ctxLogger.Info("verifying remote cache")
		remotelyCached, err := cacheableTarget.CheckRemoteCache()
		if err != nil {
			// TODO: should we force build if CheckRemoteCache fails?
			// --continue-on-pull-err
			return errors.Wrapf(err, "failed to verify remote cache %s", address)
		}

		if remotelyCached {
			ctxLogger.Info("target is remotely cached, pulling artifact")
			if pullErr := cacheableTarget.PullRemoteCache(); pullErr != nil {
				// TODO: should we force build if PullRemoteCache fails?
				// --continue-on-pull-err
				return errors.Wrap(pullErr, "the remote build cache could not be pulled")
			}
			ctxLogger.Info("successfully pulled artifacts, skipping build")
			return nil
		}

		return build()
	}
}

// String returns a string representation of the graph
func (g *Graph) String() string {
	return g.dag.StringWithNodeTypes()
}

// Dot exports the graph in dot format
func (g *Graph) Dot() []byte {
	return g.dag.Dot(&dag.DotOpts{
		Verbose:    true,
		DrawCycles: true,
		MaxDepth:   0,
	})
}

// DotToFile exports dot graph to file
func (g *Graph) DotToFile(filename string) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return nil
	}

	return ioutil.WriteFile(filename, g.Dot(), 0755)
}

// MarshalJSON marshals the graph as JSON
func (g *Graph) MarshalJSON() ([]byte, error) {
	return g.dag.MarshalJSON()
}
