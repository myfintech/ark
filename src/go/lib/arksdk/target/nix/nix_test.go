package nix

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/stretchr/testify/require"
)

func TestNix_Build(t *testing.T) {
	if nixPath, err := exec.LookPath("nix-env"); err != nil || nixPath == "" {
		t.Skip(errors.Wrap(err, "skipping nix tests"))
		return
	}

	cwd, _ := os.Getwd()
	testDataDir := filepath.Join(cwd, "testdata")

	workspace := base.NewWorkspace()
	workspace.RegisteredTargets = base.Targets{
		"nix": Target{},
	}
	require.NoError(t, workspace.DetermineRoot(testDataDir))

	diag := workspace.DecodeFile(nil)
	if diag.HasErrors() {
		require.NoError(t, diag)
	}

	buildFiles, err := workspace.DecodeBuildFiles()
	require.NoError(t, err, "no error decoding build files")
	require.NoError(t, workspace.LoadTargets(buildFiles), "must load target hcl files into workspace")

	addressable, err := workspace.TargetLUT.LookupByAddress("test.nix.test_cached")
	require.NoError(t, err, "no error looking up target address")

	target := addressable.(Target)

	require.NoError(t, target.PreBuild())

	falsy, err := target.CheckLocalBuildCache()
	require.NoError(t, err)
	require.False(t, falsy)

	require.NoError(t, target.Build())

	truthy, err := target.CheckLocalBuildCache()
	require.NoError(t, err)
	require.True(t, truthy)
}
