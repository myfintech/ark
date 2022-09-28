package probe

import (
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/dag"
)

func TestProbeTarget(t *testing.T) {
	cwd, _ := os.Getwd()
	testDataDir := filepath.Join(cwd, "testdata")

	tcpListener, err := net.Listen("tcp", "0.0.0.0:31200")
	require.NoError(t, err)
	defer func() { _ = tcpListener.Close() }()

	go func() {
		_ = http.Serve(tcpListener, http.NotFoundHandler())
	}()

	workspace := base.NewWorkspace()
	workspace.RegisteredTargets = base.Targets{
		"probe": Target{},
	}
	require.NoError(t, workspace.DetermineRoot(testDataDir), "must determine workspace root")

	diag := workspace.DecodeFile(nil)
	if diag.HasErrors() {
		require.NoError(t, diag, "must decode workspace file")
	}

	buildFiles, err := workspace.DecodeBuildFiles()
	require.NoError(t, err, "must decode build files")
	require.NoError(t, workspace.LoadTargets(buildFiles), "must load target hcl files into workspace")

	t.Run("should successfully probe tcp", func(t *testing.T) {
		target, tErr := workspace.TargetLUT.LookupByAddress("test.probe.tcp")
		require.NoError(t, tErr)

		buildWalker := base.BuildWalker(false, false, false)
		tErr = workspace.GraphWalk(target.Address(), func(vertex dag.Vertex) error {
			tErr = buildWalker(vertex)
			waitTarget := target.(Target)
			require.Equal(t, "2s", waitTarget.ComputedAttrs().Delay)
			require.Equal(t, "5s", waitTarget.ComputedAttrs().Timeout)
			require.Equal(t, 10, waitTarget.ComputedAttrs().MaxRetries)
			require.Equal(t, "tcp://0.0.0.0:31200", waitTarget.ComputedAttrs().DialAddress)
			return tErr
		})

		require.NoError(t, tErr)
	})
	t.Run("should successfully probe http with status 404", func(t *testing.T) {
		target, tErr := workspace.TargetLUT.LookupByAddress("test.probe.http")
		require.NoError(t, tErr)

		buildWalker := base.BuildWalker(false, false, false)
		tErr = workspace.GraphWalk(target.Address(), func(vertex dag.Vertex) error {
			tErr = buildWalker(vertex)
			waitTarget := target.(Target)
			require.Equal(t, "2s", waitTarget.ComputedAttrs().Delay)
			require.Equal(t, "5s", waitTarget.ComputedAttrs().Timeout)
			require.Equal(t, 10, waitTarget.ComputedAttrs().MaxRetries)
			require.Equal(t, http.StatusNotFound, waitTarget.ComputedAttrs().ExpectedStatus)
			require.Equal(t, "http://0.0.0.0:31200", waitTarget.ComputedAttrs().DialAddress)
			return tErr
		})

		require.NoError(t, tErr)
	})
}
