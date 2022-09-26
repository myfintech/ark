package base

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zclconf/go-cty/cty"

	"github.com/myfintech/ark/src/go/lib/dag"
)

type graphTarget struct {
	*RawTarget
}

func (t graphTarget) Build() error                             { return nil }
func (t graphTarget) PreBuild() error                          { return nil }
func (t graphTarget) ComputedAttrsToCty() map[string]cty.Value { return make(map[string]cty.Value) }

func TestGraph(t *testing.T) {
	graph := &Graph{}
	workspace := Workspace{}
	pkg := Package{Name: "test"}
	require.NoError(t, workspace.DetermineRootFromCWD())

	d := &graphTarget{&RawTarget{
		Type:      "test",
		Name:      "d",
		Workspace: &workspace,
		Package:   &pkg,
	}}

	c := &graphTarget{&RawTarget{
		Type:      "test",
		Name:      "c",
		Workspace: &workspace,
		Package:   &pkg,
	}}

	b := &graphTarget{&RawTarget{
		Type:      "test",
		Name:      "b",
		DependsOn: &[]string{c.Hashcode().(string)},
		Workspace: &workspace,
		Package:   &pkg,
	}}

	f := &graphTarget{&RawTarget{
		Type:      "test",
		Name:      "f",
		Workspace: &workspace,
		Package:   &pkg,
	}}

	a := &graphTarget{&RawTarget{
		Type: "test",
		Name: "a",
		DependsOn: &[]string{
			b.Hashcode().(string),
			d.Hashcode().(string),
		},
		Workspace: &workspace,
		Package:   &pkg,
	}}

	graph.Add(a)
	graph.Add(b)
	graph.Add(c)
	graph.Add(d)
	graph.Add(f)
	graph.Connect(a, b)
	graph.Connect(a, d)
	graph.Connect(b, c)
	graph.Connect(d, c)

	graph.TransitiveReduction()
	t.Log(string(graph.Dot()))

	require.NoError(t, graph.Validate())

	graph = graph.Isolate(a)
	err := graph.Walk(func(vertex dag.Vertex) error {
		target := vertex.(Buildable)
		t.Log("build", target)
		require.NoError(t, target.PreBuild())
		require.NoError(t, target.Build())
		return nil
	})
	require.NoError(t, err)

	expectedVertices := graph.TopologicalSort(a)
	// Ensures the expected order was received from the sort
	require.Equal(t, expectedVertices, []dag.Vertex{c, b, d, a})

	for _, vertex := range expectedVertices {
		t.Log(vertex)
	}

	// verifies this function can be called n times and the result is deterministic
	for i := 0; i < 100; i++ {
		require.Equal(t, expectedVertices, graph.TopologicalSort(a))
	}
}
