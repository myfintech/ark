package jsonnet_file

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

		attrs := target.ComputedAttrs()
		require.NotEmpty(t, attrs.File)
		require.NotEmpty(t, attrs.Variables)
		require.Equal(t, "json", attrs.Format)
		require.NoError(t, os.RemoveAll(target.RenderedFilePath()))

		return nil
	})
}
