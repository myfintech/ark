package module

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/build"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/stretchr/testify/require"
)

// Load workspace from testdata_modules
// Decode workspace and build module target

func TestModules(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	testDataDir := filepath.Join(cwd, "testdata")

	workspace := base.NewWorkspace()

	require.NoError(t, workspace.DetermineRoot(testDataDir))

	workspace.RegisteredTargets = base.Targets{
		"build": build.Target{},
	}

	err = workspace.InitDockerClient()
	require.NoError(t, err)

	buildFiles, decodeErr := workspace.DecodeBuildFiles()
	require.NoError(t, decodeErr)

	err = workspace.LoadTargets(buildFiles)
	require.NoError(t, err)

	vertex, vertErr := workspace.TargetLUT.LookupByAddress("test.mod.build.test")
	require.NoError(t, vertErr)

	target := vertex.(build.Target)

	err = target.PreBuild()
	require.NoError(t, err)

	err = target.Build()
	require.NoError(t, err)
}
