package base

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTargetLUT(t *testing.T) {
	pkg := Package{Name: "test"}
	workspace := Workspace{}
	targetLUT := LookupTable{}

	require.NoError(t, workspace.DetermineRootFromCWD())

	a := &graphTarget{&RawTarget{
		Type:      "go_binary",
		Name:      "linux",
		Workspace: &workspace,
		Package:   &pkg,
	}}

	t.Run("should be able to add a target to the table", func(t *testing.T) {
		require.NoError(t, targetLUT.Add(a))
		require.Equal(t, a, targetLUT[a.Address()])
	})

	t.Run("should be able to lookup a target by its package, type and name", func(t *testing.T) {
		b, err := targetLUT.Lookup(a.Package.Name, a.Type, a.Name)
		require.NoError(t, err)
		require.Equal(t, a, b)
	})

	t.Run("should throw an error if a target doesn't exist", func(t *testing.T) {
		_, err := targetLUT.Lookup("", "", "")
		require.Error(t, err)
	})
}
