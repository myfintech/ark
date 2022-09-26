package jsonnet

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/dag"
)

func TestJsonnetTarget_Build(t *testing.T) {
	cwd, _ := os.Getwd()
	testDataDir := filepath.Join(cwd, "testdata")

	workspace := base.NewWorkspace()
	workspace.RegisteredTargets = base.Targets{
		"jsonnet": Target{},
	}
	require.NoError(t, workspace.DetermineRoot(testDataDir))

	buildFiles, err := workspace.DecodeBuildFiles()
	require.NoError(t, err, "no error decoding build files")

	require.NoError(t, workspace.LoadTargets(buildFiles), "must load target hcl files into workspace")

	intendedTarget, err := workspace.TargetLUT.LookupByAddress("test.jsonnet.test")
	require.NoError(t, err)

	_ = workspace.GraphWalk(intendedTarget.Address(), func(vertex dag.Vertex) error {
		buildable := vertex.(base.Buildable)
		require.NoError(t, buildable.PreBuild())
		require.NoError(t, buildable.Build())

		target := buildable.(Target)
		require.True(t, *target.YamlOut)

		attrs := target.ComputedAttrs()
		require.NotEmpty(t, attrs.OutputDir)
		require.NotEmpty(t, attrs.Files)
		require.NotEmpty(t, attrs.Variables)

		for _, file := range attrs.Files {
			require.NoError(t, os.RemoveAll(target.ConstructOutFilePath(file)))
		}
		return nil
	})
}
