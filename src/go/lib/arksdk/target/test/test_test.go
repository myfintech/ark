package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/dag"
	"github.com/myfintech/ark/src/go/lib/utils"
	"github.com/stretchr/testify/require"
)

func TestTest(t *testing.T) {
	cwd, _ := os.Getwd()
	testDataDir := filepath.Join(cwd, "testdata")

	workspace := base.NewWorkspace()

	workspace.RegisteredTargets = base.Targets{
		"test": Target{},
	}

	require.NoError(t, workspace.DetermineRoot(testDataDir))

	diag := workspace.DecodeFile(nil)

	if diag != nil && diag.HasErrors() {
		require.NoError(t, diag)
	}

	require.NoError(t, workspace.InitKubeClient(utils.EnvLookup("ARK_K8S_NAMESPACE", "default")))
	require.NoError(t, workspace.InitDockerClient())

	buildFiles, err := workspace.DecodeBuildFiles()
	require.NoError(t, err, "no error decoding build files")

	require.NoError(t, workspace.LoadTargets(buildFiles), "must load target hcl files into workspace")

	currentContext, err := workspace.K8s.CurrentContext()
	require.NoError(t, err)

	if !utils.IsK8sContextSafe([]string{"docker-desktop", "development_sre"}, "ARK_K8S_SAFE_CONTEXTS", currentContext) {
		t.Skip("Skipping test because context is not designated as safe")
		return
	}

	t.Run("basic success", func(t *testing.T) {
		require.NoError(t, walkByTarget(t, workspace, "test.test.test1"))
	})
	t.Run("basic fail", func(t *testing.T) {
		require.Error(t, walkByTarget(t, workspace, "test.test.test2"))
	})
}

func walkByTarget(t *testing.T, workspace *base.Workspace, address string) error {
	return workspace.GraphWalk(address, func(vertex dag.Vertex) error {
		buildable := vertex.(base.Buildable)
		_, cacheable := buildable.(base.Cacheable)
		require.Equal(t, true, cacheable)

		require.NoError(t, buildable.PreBuild())

		if buildErr := buildable.Build(); buildErr != nil {
			return buildErr
		}
		return nil
	})
}
