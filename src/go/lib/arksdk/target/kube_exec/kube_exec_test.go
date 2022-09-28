package kube_exec

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/myfintech/ark/src/go/lib/utils"

	"github.com/stretchr/testify/require"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/dag"
)

func TestKubeExec_Build(t *testing.T) {
	// FIXME refactor this test to deploy and tear down its own resource for the test
	t.Skip("skipping test because it should not rely on a deployed resource")

	cwd, _ := os.Getwd()
	testDataDir := filepath.Join(cwd, "testdata")

	workspace := base.NewWorkspace()
	workspace.RegisteredTargets = base.Targets{
		"kube_exec": Target{},
	}
	require.NoError(t, workspace.DetermineRoot(testDataDir))

	diag := workspace.DecodeFile(nil)
	if diag.HasErrors() {
		require.NoError(t, diag)
	}

	buildFiles, err := workspace.DecodeBuildFiles()
	require.NoError(t, err, "no error decoding build files")

	require.NoError(t, workspace.LoadTargets(buildFiles), "must load target hcl files into workspace")

	require.NoError(t, workspace.InitKubeClient(utils.EnvLookup("ARK_K8S_NAMESPACE", "default")), "must initiate kube client")

	workspace.K8s.NamespaceOverride = "nginx-ingress"

	currentContext, _ := workspace.K8s.CurrentContext()
	if !utils.IsK8sContextSafe([]string{"development_sre"}, "ARK_K8S_SAFE_CONTEXTS", currentContext) {
		t.Skip("Skipping test because context is not designated as safe")
		return
	}

	t.Run("Command Success", func(t *testing.T) {
		require.NoError(t, walkByTarget(t, workspace, "test.kube_exec.test"))
	})

	t.Run("Command Failure", func(t *testing.T) {
		require.Error(t, walkByTarget(t, workspace, "test.kube_exec.fail_test"))
	})
}

func walkByTarget(t *testing.T, workspace *base.Workspace, address string) error {
	intendedTarget, err := workspace.TargetLUT.LookupByAddress(address)
	if err != nil {
		return err
	}

	return workspace.GraphWalk(intendedTarget.Address(), func(vertex dag.Vertex) error {
		buildable := vertex.(base.Buildable)
		if preBuildErr := buildable.PreBuild(); preBuildErr != nil {
			return preBuildErr
		}
		if buildErr := buildable.Build(); buildErr != nil {
			return buildErr
		}

		execTarget := buildable.(Target)
		_, cacheable := buildable.(base.Cacheable)
		require.Equal(t, false, cacheable)
		require.NotEmpty(t, execTarget.ComputedAttrs().Command)
		require.NotEmpty(t, execTarget.Command, "GetStateAttrs.Command should not be empty.")
		return nil
	})
}
