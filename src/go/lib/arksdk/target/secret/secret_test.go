package secret

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/myfintech/ark/src/go/lib/utils"

	"github.com/myfintech/ark/src/go/lib/dag"
	"github.com/myfintech/ark/src/go/lib/kube"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/stretchr/testify/require"
)

func TestSecretTarget(t *testing.T) {
	cwd, _ := os.Getwd()
	testDataDir := filepath.Join(cwd, "testdata")

	workspace := base.NewWorkspace()
	workspace.RegisteredTargets = base.Targets{
		"secret": Target{},
	}
	require.NoError(t, workspace.DetermineRoot(testDataDir), "must determine workspace root")

	diag := workspace.DecodeFile(nil)
	if diag.HasErrors() {
		require.NoError(t, diag, "must decode workspace file")
	}

	require.NoError(t, workspace.InitKubeClient(utils.EnvLookup("ARK_K8S_NAMESPACE", "default")))

	buildFiles, err := workspace.DecodeBuildFiles()
	require.NoError(t, err, "must decode build files")

	require.NoError(t, workspace.LoadTargets(buildFiles), "must load target hcl files into workspace")

	currentContext, err := workspace.K8s.CurrentContext()
	require.NoError(t, err)

	if !utils.IsK8sContextSafe([]string{"docker-desktop", "development_sre"}, "ARK_K8S_SAFE_CONTEXTS", currentContext) {
		t.Skip("Skipping test because context is not designated as safe")
		return
	}

	t.Run("target with optional files present", func(t *testing.T) {
		require.NoError(t, walkByTarget(t, workspace, "test.secret.test_1"))
	})

	t.Run("target with optional files missing", func(t *testing.T) {
		require.NoError(t, walkByTarget(t, workspace, "test.secret.test_2"))
	})

	t.Run("target with required files present", func(t *testing.T) {
		require.NoError(t, walkByTarget(t, workspace, "test.secret.test_3"))
	})

	t.Run("target with required files missing", func(t *testing.T) {
		require.Error(t, walkByTarget(t, workspace, "test.secret.test_4"))
	})

	t.Run("target with optional env vars present", func(t *testing.T) {
		require.NoError(t, walkByTarget(t, workspace, "test.secret.test_5"))
	})

	t.Run("target with optional env vars missing", func(t *testing.T) {
		require.NoError(t, walkByTarget(t, workspace, "test.secret.test_6"))
	})

	t.Run("target with required env vars present", func(t *testing.T) {
		require.NoError(t, walkByTarget(t, workspace, "test.secret.test_7"))
	})

	t.Run("target with required env vars missing", func(t *testing.T) {
		require.Error(t, walkByTarget(t, workspace, "test.secret.test_8"))
	})

	t.Run("target with neither files nor env vars declared", func(t *testing.T) {
		require.Error(t, walkByTarget(t, workspace, "test.secret.test_9"))
	})

	t.Run("target with both files and env vars declared", func(t *testing.T) {
		require.Error(t, walkByTarget(t, workspace, "test.secret.test_10"))
	})

	t.Run("target with a directory as files value", func(t *testing.T) {
		require.NoError(t, walkByTarget(t, workspace, "test.secret.test_11"))
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

		target := buildable.(Target)
		attrs := target.ComputedAttrs()
		restClient, clientErr := workspace.K8s.Factory.RESTClient()
		if clientErr != nil {
			return clientErr
		}
		_, getErr := kube.GetSecret(restClient, workspace.K8s.Namespace(), attrs.SecretName)
		if getErr != nil && !attrs.Optional {
			return getErr
		}
		if getErr == nil {
			exists, cacheErr := target.CheckLocalBuildCache()
			if cacheErr != nil {
				return cacheErr
			}
			require.True(t, exists)
			if deleteErr := kube.DeleteSecret(restClient, workspace.K8s.Namespace(), attrs.SecretName); deleteErr != nil {
				return deleteErr
			}
		}

		return nil
	})
}
