package http_archive

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/dag"
)

func TestHTTPArchiveTarget_Build(t *testing.T) {
	cwd, _ := os.Getwd()
	testDataDir := filepath.Join(cwd, "testdata")

	defer func() {
		_ = os.RemoveAll("testdata/.ark")
	}()

	workspace := base.NewWorkspace()
	workspace.RegisteredTargets = base.Targets{
		"http_archive": Target{},
	}
	require.NoError(t, workspace.DetermineRoot(testDataDir))

	buildFiles, err := workspace.DecodeBuildFiles()
	require.NoError(t, err, "no error decoding build files")

	require.NoError(t, workspace.LoadTargets(buildFiles), "must load target hcl files into workspace")

	intendedTarget, err := workspace.TargetLUT.LookupByAddress("test.http_archive.test")
	require.NoError(t, err)

	_ = workspace.GraphWalk(intendedTarget.Address(), func(vertex dag.Vertex) error {
		buildable := vertex.(base.Buildable)

		require.NoError(t, buildable.PreBuild())
		require.NoError(t, buildable.Build())

		return nil
	})
}
