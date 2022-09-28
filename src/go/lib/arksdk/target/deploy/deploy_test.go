package deploy

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/myfintech/ark/src/go/lib/kube"
	"github.com/myfintech/ark/src/go/lib/utils"

	"github.com/stretchr/testify/require"

	"github.com/myfintech/ark/src/go/lib/dag"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
)

func TestKubernetesDeploy(t *testing.T) {
	cwd, _ := os.Getwd()
	testDataDir := filepath.Join(cwd, "testdata")

	workspace := base.NewWorkspace()

	workspace.RegisteredTargets = base.Targets{
		"deploy": Target{},
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

	t.Run("sync server disabled", func(t *testing.T) {
		require.NoError(t, walkByTarget(t, workspace, "test.deploy.test1"))
	})
}

func walkByTarget(t *testing.T, workspace *base.Workspace, address string) error {
	return workspace.GraphWalk(address, func(vertex dag.Vertex) error {
		buildable := vertex.(base.Buildable)
		if preBuildErr := buildable.PreBuild(); preBuildErr != nil {
			return preBuildErr
		}
		if buildErr := buildable.Build(); buildErr != nil {
			return buildErr
		}

		target := buildable.(Target)
		_, cacheable := buildable.(base.Cacheable)
		require.Equal(t, false, cacheable)
		t.Log(target.RenderedFilePath())

		attrs := target.ComputedAttrs()

		for _, action := range attrs.LiveSyncOnActions {
			require.Equal(t, []string{"yarn"}, action.Command)
			t.Log(utils.MarshalJSONSafe(attrs.LiveSyncOnActions, true))
		}

		if deleteErr := kube.Delete(workspace.K8s, workspace.K8s.Namespace(), 10*time.Second, target.RenderedFilePath()); deleteErr != nil {
			return deleteErr
		}
		return nil
	})
}
