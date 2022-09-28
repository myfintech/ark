package watchman

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/myfintech/ark/src/go/lib/utils"
)

func skipIfWatchmanSocketMounted(t *testing.T, run func(*testing.T)) {
	socketOverride := os.Getenv("WATCHMAN_SOCK")
	if socketOverride != "" {
		t.Skip("skipping test. watchman socket is mounted directly")
		return
	}
	run(t)
}

func TestPath(t *testing.T) {
	skipIfWatchmanSocketMounted(t, func(t *testing.T) {
		watchmanPath, err := Path()
		require.NoError(t, err)
		require.NotEmpty(t, watchmanPath)
	})
}

func TestIsBinaryInstalled(t *testing.T) {
	skipIfWatchmanSocketMounted(t, func(t *testing.T) {
		require.True(t, IsBinaryInstalled())
	})
}

func TestGetSocketName(t *testing.T) {
	skipIfWatchmanSocketMounted(t, func(t *testing.T) {
		socketName, err := GetSocketName()
		require.NoError(t, err)
		require.NotEmpty(t, socketName)
	})
}

func TestJSONUnmarshallCommand_GenericError(t *testing.T) {
	skipIfWatchmanSocketMounted(t, func(t *testing.T) {
		cmd := exec.Command("watchman", "-j")
		cmd.Stdin = bytes.NewBufferString("Garbage that watchman won't understand")
		require.Error(t, JSONUnmarshalCommand(cmd, &map[string]interface{}{}), "should raise error when we exit with non-zero status")
	})
}

func TestJSONUnmarshallCommand_RaiseOnErrorKey(t *testing.T) {
	skipIfWatchmanSocketMounted(t, func(t *testing.T) {
		cmd := exec.Command("watchman", "-j")
		cmd.Stdin = bytes.NewBufferString(utils.MarshalJSONSafe([]interface{}{
			"query", "/A/World/That/Doesnt/Exist", map[string]interface{}{},
		}, false))
		require.Error(t, JSONUnmarshalCommand(cmd, &map[string]interface{}{}), "the command should return a JSON object with an error key containing a string")
	})
}

func TestWatchmanIntegration(t *testing.T) {
	wd, _ := os.Getwd()
	client, connErr := Connect(context.Background(), 30)
	require.NoError(t, connErr)
	defer func() { _ = client.Close() }()

	t.Run("should be able to list capabilities", func(t *testing.T) {
		resp, err := client.ListCapabilities()
		require.NoError(t, err)
		require.NotEmpty(t, resp.Capabilities)
		require.Equal(t, resp.HasCapability("cmd-watch-project"), true)
		require.Equal(t, resp.HasCapability("nope"), false)
	})

	t.Run("should fail capabilities check", func(t *testing.T) {
		_, err := client.CheckCapabilities(VersionOptions{
			Required: []string{"fake-capability"},
		})
		require.Error(t, err)
	})

	t.Run("should pass capabilities check", func(t *testing.T) {
		requiredCmd := "cmd-watch-project"
		version, err := client.CheckCapabilities(VersionOptions{
			Required: []string{requiredCmd},
		})
		require.NoError(t, err)
		require.NotEmpty(t, version.Version)
		require.Equal(t, version.Capabilities[requiredCmd], true)
	})

	t.Run("should be able to execute the find after watch", func(t *testing.T) {
		_, err := client.Watch(WatchOptions{Directory: wd})
		fileResp, err := client.Find(FindOptions{
			Directory: wd,
			Patterns:  []string{"*.go"},
		})
		require.NoError(t, err)
		require.NotEmpty(t, fileResp.Files)
	})

	t.Run("should be able to subscribe to file system changes", func(t *testing.T) {
		sub, err := Subscribe(SubscribeOptions{
			Name: "test",
			Root: wd,
			Filter: &QueryFilter{
				// Since:      "c:1582404078:98505:4:434",
				Fields:     []string{"name", "cclock"},
				Expression: []interface{}{"match", "*.go"},
				DeferVcs:   true,
				Defer:      nil,
				Drop:       nil,
			},
		}, false, false)
		require.NoError(t, err)

		// ignore initial state
		_ = <-sub.ChangeFeed

		// touch a local file to trigger a change
		_ = exec.Command("touch", filepath.Join(wd, "watchman_test.go")).Run()

		// catch that change and verify it was the correct file
		change := <-sub.ChangeFeed
		require.Equal(t, change.Files[0].Name, "watchman_test.go")
	})

}
