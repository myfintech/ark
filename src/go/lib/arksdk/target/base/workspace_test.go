package base

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zclconf/go-cty/cty"
)

type exampleSetupTarget struct {
	*RawTarget
	Special string `hcl:"special,attr"`
}

// Build
func (t exampleSetupTarget) Build() error                             { return nil }
func (t exampleSetupTarget) PreBuild() error                          { return nil }
func (t exampleSetupTarget) ComputedAttrsToCty() map[string]cty.Value { return nil }

// TODO: Add a validation case for ensuring there is only one WORKSPACE.hcl file
func TestWorkspace(t *testing.T) {
	workspace := NewWorkspace()
	cwd, cwdErr := os.Getwd()
	require.NoError(t, cwdErr)

	testWSDir := filepath.Join(cwd, "testdata")

	t.Run("workspace should be able to locate it's root from cwd", func(t *testing.T) {
		require.NoError(t, os.Chdir(testWSDir))
		defer func() { _ = os.Chdir(cwd) }()
		require.NoError(t, workspace.DetermineRootFromCWD(), "should not fail to locate workspace root")
		require.NotEmpty(t, workspace.Dir)
		require.Equal(t, filepath.Join(workspace.Dir, "WORKSPACE.hcl"), workspace.File)
		require.Equal(t, filepath.Join(testWSDir, "WORKSPACE.hcl"), workspace.File)
	})

	t.Run("workspace should be able to decode its WORKSPACE.hcl file", func(t *testing.T) {
		diag := workspace.DecodeFile(nil)
		if diag.HasErrors() {
			require.NoError(t, diag)
		}
		require.Equal(t, "gs://ark-cache", workspace.Config.Artifacts.StorageBaseURL)
		require.Len(t, workspace.Config.Plugins, 2)
	})

	t.Run("should be able to walk the workspace root for build files", func(t *testing.T) {
		buildFiles, err := workspace.LoadBuildFiles()
		require.NoError(t, err)
		require.NotEmpty(t, buildFiles)
		require.NotContains(t, buildFiles, filepath.Join(testWSDir, "sub", "BUILD.hcl"), "should not contain sub workspace BUILD.hcl file")
	})

	t.Run("should be able to set configuration environment on workspace or default to sane value", func(t *testing.T) {
		workspace.SetEnvironmentConstraint("remote")
		require.Equal(t, "remote", workspace.ConfigurationEnvironment)
		workspace.SetEnvironmentConstraint("")
		require.Equal(t, "local", workspace.ConfigurationEnvironment)
	})
}

func TestWorkspace_LoadTargets(t *testing.T) {
	cwd, _ := os.Getwd()
	testDataDir := filepath.Join(cwd, "testdata_setup")

	workspace := NewWorkspace()
	workspace.RegisteredTargets = Targets{
		"arktest": exampleSetupTarget{},
	}
	require.NoError(t, workspace.DetermineRoot(testDataDir))

	buildFiles, err := workspace.DecodeBuildFiles()
	require.NoError(t, err)

	require.NoError(t, workspace.LoadTargets(buildFiles))

	target, err := workspace.TargetLUT.Lookup("setup", "arktest", "example")
	require.NoError(t, err, "should be able to locate target by package, type, and name.")

	example := target.(exampleSetupTarget)

	require.Equal(t, "testing", example.Special)

}
