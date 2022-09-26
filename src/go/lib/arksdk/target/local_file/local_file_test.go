package local_file

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/dag"
)

func TestLocalExecTarget(t *testing.T) {
	cwd, _ := os.Getwd()
	testDataDir := filepath.Join(cwd, "testdata")

	workspace := base.NewWorkspace()
	workspace.RegisteredTargets = base.Targets{
		"local_file": Target{},
	}
	require.NoError(t, workspace.DetermineRoot(testDataDir), "must determine workspace root")

	diag := workspace.DecodeFile(nil)
	if diag.HasErrors() {
		require.NoError(t, diag, "must decode workspace file")
	}
	require.Equal(t, "gcr.io/managed-infrastructure/ark/plugin-test:latest", workspace.Config.Plugins[0].Image)

	buildFiles, err := workspace.DecodeBuildFiles()
	require.NoError(t, err, "must decode build files")

	require.NoError(t, workspace.LoadTargets(buildFiles), "must load target hcl files into workspace")

	t.Run("regular local file good", func(t *testing.T) {
		require.NoError(t, walkByTarget(t, workspace, "test.local_file.test1", true), "build should succeed")
	})

	t.Run("regular local file absolute good", func(t *testing.T) {
		require.NoError(t, walkByTarget(t, workspace, "test.local_file.test2", true), "build should succeed")
	})

	t.Run("plugin local file good", func(t *testing.T) {
		require.NoError(t, walkByTarget(t, workspace, "test.local_file.test3", true), "build should succeed")
	})
}

func walkByTarget(t *testing.T, workspace *base.Workspace, address string, successCase bool) error {
	intendedTarget, err := workspace.TargetLUT.LookupByAddress(address)
	if err != nil {
		return err
	}

	return workspace.GraphWalk(intendedTarget.Address(), func(vertex dag.Vertex) error {
		buildable := vertex.(base.Buildable)
		fileTarget := buildable.(Target)
		attrs := fileTarget.ComputedAttrs()

		cached, cacheErr := fileTarget.CheckLocalBuildCache()
		if cacheErr != nil {
			return cacheErr
		}
		require.False(t, cached, "artifact should not be cached")

		if preBuildErr := fileTarget.PreBuild(); preBuildErr != nil {
			return preBuildErr
		}
		if buildErr := fileTarget.Build(); buildErr != nil {
			return buildErr
		}

		if successCase {
			content, readErr := ioutil.ReadFile(fileTarget.Artifact())
			if readErr != nil {
				return readErr
			}
			require.Equal(t, attrs.Content, string(content), "file contents should match content attribute")
		}

		if saveStateErr := fileTarget.RawTarget.SaveLocalBuildCacheState(); saveStateErr != nil {
			return saveStateErr
		}
		cached, cacheErr = fileTarget.CheckLocalBuildCache()
		if cacheErr != nil {
			return cacheErr
		}
		require.True(t, cached, "artifact should be cached")

		if removeErr := os.RemoveAll(filepath.Join(fileTarget.Workspace.ArtifactsDir(), fileTarget.Address())); removeErr != nil {
			return removeErr
		}
		return nil
	})
}
