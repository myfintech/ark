package dag

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"

	"github.com/hashicorp/terraform/tfdiags"

	"github.com/hashicorp/go-multierror"

	"github.com/myfintech/ark/src/go/lib/log"
)

var ctxLog = log.WithFields(log.Fields{
	"prefix": "dag",
})

// AcyclicGraph is a specialization of Graph that cannot have cycles. With
// this property, we get the property of sane graph traversal.
type AcyclicGraph struct {
	Graph
}

// WalkFunc is the callback used for walking the graph.
type WalkFunc func(Vertex) tfdiags.Diagnostics

// WalkFuncWithErr an alias for WalkFunc that returns standard go errors instead of terraform diagnostics
type WalkFuncWithErr func(Vertex) error

// DepthWalkFunc is a walk function that also receives the current depth of the
// walk as an argument
type DepthWalkFunc func(Vertex, int) error

func (g *AcyclicGraph) DirectedGraph() Grapher {
	return g
}

// Ancestors returns a Set that includes every Vertex yielded by walking down from the
// provided starting Vertex v.
func (g *AcyclicGraph) Ancestors(v Vertex) (*Set, error) {
	s := new(Set)
	start := AsVertexList(g.DownEdges(v))
	memoFunc := func(v Vertex, d int) error {
		s.Add(v)
		return nil
	}

	if err := g.DepthFirstWalk(start, memoFunc); err != nil {
		return nil, err
	}

	return s, nil
}

// Descendents returns a Set that includes every Vertex yielded by walking up from the
// provided starting Vertex v.
func (g *AcyclicGraph) Descendents(v Vertex) (*Set, error) {
	s := new(Set)
	start := AsVertexList(g.UpEdges(v))
	memoFunc := func(v Vertex, d int) error {
		s.Add(v)
		return nil
	}

	if err := g.ReverseDepthFirstWalk(start, memoFunc); err != nil {
		return nil, err
	}

	return s, nil
}

// Root returns the root of the DAG, or an error.
//
// Complexity: O(V)
func (g *AcyclicGraph) Root() (Vertex, error) {
	roots := make([]Vertex, 0, 1)
	for _, v := range g.Vertices() {
		if g.UpEdges(v).Len() == 0 {
			roots = append(roots, v)
		}
	}

	if len(roots) > 1 {
		// TODO(mitchellh): make this error message a lot better
		return nil, fmt.Errorf("multiple roots: %#v", roots)
	}

	if len(roots) == 0 {
		return nil, fmt.Errorf("no roots found")
	}

	return roots[0], nil
}

// TransitiveReduction performs the transitive reduction of graph g in place.
// The transitive reduction of a graph is a graph with as few edges as
// possible with the same reachability as the original graph. This means
// that if there are three nodes A => B => C, and A connects to both
// B and C, and B connects to C, then the transitive reduction is the
// same graph with only a single edge between A and B, and a single edge
// between B and C.
//
// The graph must be valid for this operation to behave properly. If
// Validate() returns an error, the behavior is undefined and the results
// will likely be unexpected.
//
// Complexity: O(V(V+E)), or asymptotically O(VE)
func (g *AcyclicGraph) TransitiveReduction() {
	// For each vertex u in graph g, do a DFS starting from each vertex
	// v such that the edge (u,v) exists (v is a direct descendant of u).
	//
	// For each v-prime reachable from v, remove the edge (u, v-prime).
	defer g.debug.BeginOperation("TransitiveReduction", "").End("")

	for _, u := range g.Vertices() {
		uTargets := g.DownEdges(u)
		vs := AsVertexList(g.DownEdges(u))

		_ = g.DepthFirstWalkWithOptionalSort(vs, false, func(v Vertex, d int) error {
			shared := uTargets.Intersection(g.DownEdges(v))
			for _, vPrime := range AsVertexList(shared) {
				g.RemoveEdge(BasicEdge(u, vPrime))
			}

			return nil
		})
	}
}

// Validate validates the DAG. A DAG is valid if it has a single root
// with no cycles.
func (g *AcyclicGraph) Validate() error {
	if _, err := g.Root(); err != nil {
		return err
	}

	// Look for cycles of more than 1 component
	var err error
	cycles := g.Cycles()
	if len(cycles) > 0 {
		for _, cycle := range cycles {
			cycleStr := make([]string, len(cycle))
			for j, vertex := range cycle {
				cycleStr[j] = VertexName(vertex)
			}

			err = multierror.Append(err, fmt.Errorf(
				"Cycle: %s", strings.Join(cycleStr, ", ")))
		}
	}

	// Look for cycles to self
	for _, e := range g.Edges() {
		if e.Source() == e.Target() {
			err = multierror.Append(err, fmt.Errorf(
				"Self reference: %s", VertexName(e.Source())))
		}
	}

	return err
}

// Cycles detects a strongly connected graph and returns a set of cycles if present
func (g *AcyclicGraph) Cycles() [][]Vertex {
	var cycles [][]Vertex
	for _, cycle := range StronglyConnected(&g.Graph) {
		if len(cycle) > 1 {
			cycles = append(cycles, cycle)
		}
	}
	return cycles
}

// Walk walks the graph, calling your callback as each node is visited.
// This will walk nodes in parallel if it can. The resulting diagnostics
// contains problems from all graphs visited, in no particular order.
func (g *AcyclicGraph) Walk(cb WalkFunc) tfdiags.Diagnostics {
	defer g.debug.BeginOperation(typeWalk, "").End("")

	w := &Walker{Callback: cb, Reverse: true}
	w.Update(g)
	return w.Wait()
}

// WalkWithErr an idiomatic wrapper around the terraform walk func that uses diagnostic errors
func (g AcyclicGraph) WalkWithErr(walk WalkFuncWithErr) error {
	return g.Walk(func(vertex Vertex) tfdiags.Diagnostics {
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

func AsVertexList(s *Set) []Vertex {
	rawList := s.List()
	vertexList := make([]Vertex, len(rawList))
	for i, raw := range rawList {
		vertexList[i] = raw.(Vertex)
	}
	return vertexList
}

type vertexAtDepth struct {
	Vertex Vertex
	Depth  int
}

// DepthFirstWalkWithOptionalSort does a depth-first walk of the graph starting from
// the vertices in start.
func (g *AcyclicGraph) DepthFirstWalk(start []Vertex, f DepthWalkFunc) error {
	return g.DepthFirstWalkWithOptionalSort(start, true, f)
}

// DepthFirstWalkWithOptionalSort
// This internal method provides the option of not sorting the vertices during
// the walk, which we use for the Transitive reduction.
// Some configurations can lead to fully-connected subgraphs, which makes our
// transitive reduction algorithm O(n^3). This is still passable for the size
// of our graphs, but the additional n^2 sort operations would make this
// uncomputable in a reasonable amount of time.
func (g *AcyclicGraph) DepthFirstWalkWithOptionalSort(start []Vertex, sorted bool, f DepthWalkFunc) error {
	defer g.debug.BeginOperation(typeDepthFirstWalk, "").End("")

	seen := make(map[Vertex]struct{})
	frontier := make([]*vertexAtDepth, len(start))
	for i, v := range start {
		frontier[i] = &vertexAtDepth{
			Vertex: v,
			Depth:  0,
		}
	}
	for len(frontier) > 0 {
		// Pop the current vertex
		n := len(frontier)
		current := frontier[n-1]
		frontier = frontier[:n-1]

		// Check if we've seen this already and return...
		if _, ok := seen[hashcode(current.Vertex)]; ok {
			continue
		}

		seen[hashcode(current.Vertex)] = struct{}{}

		// Visit the current node
		if err := f(current.Vertex, current.Depth); err != nil {
			return err
		}

		// Visit targets of this in a consistent order.
		targets := AsVertexList(g.DownEdges(current.Vertex))

		if sorted {
			sort.Sort(ByVertexName(targets))
		}

		for _, t := range targets {
			frontier = append(frontier, &vertexAtDepth{
				Vertex: t,
				Depth:  current.Depth + 1,
			})
		}
	}

	return nil
}

// TopologicalSort of a directed graph is a linear ordering of its vertices such that
// for every directed edge uv from vertex u to vertex v, u comes before v in the ordering.
// This function sorts the vertices by their name to be deterministic.
func (g *AcyclicGraph) TopologicalSort(start Vertex) []Vertex {
	defer g.debug.BeginOperation(typeDepthFirstWalk, "").End("")

	path := []Vertex{start}
	seen := make(map[Vertex]int)
	frontier := make([]*vertexAtDepth, len(path))

	for i, v := range path {
		frontier[i] = &vertexAtDepth{
			Vertex: v,
			Depth:  0,
		}
	}

	for len(frontier) > 0 {
		// Pop the current vertex
		n := len(frontier)
		current := frontier[n-1]
		frontier = frontier[:n-1]

		// Check if we've seen this already and updates its position in the path
		if idx, ok := seen[hashcode(current.Vertex)]; ok {
			// remove the element from the slice
			path = append(path[:idx], path[idx+1:]...)
		}

		// skip the starting vertex
		if hashcode(current.Vertex) != hashcode(start) {
			path = append(path, current.Vertex)
		}

		seen[hashcode(current.Vertex)] = len(path) - 1

		// Visit targets of this in a consistent order.
		targets := AsVertexList(g.DownEdges(current.Vertex))
		sort.Sort(ByVertexName(targets))

		for _, t := range targets {
			frontier = append(frontier, &vertexAtDepth{
				Vertex: t,
				Depth:  current.Depth + 1,
			})
		}
	}

	// reverse the order of the slice
	for i := len(path)/2 - 1; i >= 0; i-- {
		opp := len(path) - 1 - i
		path[i], path[opp] = path[opp], path[i]
	}

	return path
}

// Isolate returns an isolated sub graph from the starting vertex
func (g *AcyclicGraph) Isolate(start Vertex) *AcyclicGraph {
	graph := new(AcyclicGraph)
	nodes := make([]Vertex, 0)
	_ = g.DepthFirstWalk([]Vertex{start}, func(vertex Vertex, i int) error {
		nodes = append(nodes, vertex)
		return nil
	})
	for _, vertex := range nodes {
		graph.Add(vertex)
		for _, edge := range g.EdgesFrom(vertex) {
			graph.Add(edge.Source())
			graph.Add(edge.Target())
			graph.Connect(BasicEdge(edge.Source(), edge.Target()))
		}
	}
	graph.TransitiveReduction()
	return graph
}

// ReverseDepthFirstWalk does a depth-first walk _up_ the graph starting from
// the vertices in start.
func (g *AcyclicGraph) ReverseDepthFirstWalk(start []Vertex, f DepthWalkFunc) error {
	defer g.debug.BeginOperation(typeReverseDepthFirstWalk, "").End("")

	seen := make(map[Vertex]struct{})
	frontier := make([]*vertexAtDepth, len(start))
	for i, v := range start {
		frontier[i] = &vertexAtDepth{
			Vertex: v,
			Depth:  0,
		}
	}
	for len(frontier) > 0 {
		// Pop the current vertex
		n := len(frontier)
		current := frontier[n-1]
		frontier = frontier[:n-1]

		// Check if we've seen this already and return...
		if _, ok := seen[hashcode(current.Vertex)]; ok {
			continue
		}
		seen[hashcode(current.Vertex)] = struct{}{}

		// Add next set of targets in a consistent order.
		targets := AsVertexList(g.UpEdges(current.Vertex))
		sort.Sort(ByVertexName(targets))
		for _, t := range targets {
			frontier = append(frontier, &vertexAtDepth{
				Vertex: t,
				Depth:  current.Depth + 1,
			})
		}

		// Visit the current node
		if err := f(current.Vertex, current.Depth); err != nil {
			return err
		}
	}

	return nil
}

// ByVertexName implements sort.Interface so a list of Vertices can be sorted
// consistently by their VertexName
type ByVertexName []Vertex

func (b ByVertexName) Len() int      { return len(b) }
func (b ByVertexName) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b ByVertexName) Less(i, j int) bool {
	return VertexName(b[i]) < VertexName(b[j])
}