package http_server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/messages"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs/protocols/watermill/gochannel"

	"github.com/myfintech/ark/src/go/lib/logz"

	"golang.org/x/sync/errgroup"

	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/ark/storage/memory"
	"github.com/myfintech/ark/src/go/lib/ark/targets/docker_image"
	"github.com/myfintech/ark/src/go/lib/utils"
	"github.com/stretchr/testify/require"
)

func TestHostServerWithClient(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	testdata := filepath.Join(cwd, "testdata")

	err = os.MkdirAll(testdata, 0755)
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(testdata)
	}()

	stores := []ark.Store{
		new(memory.Store),
	}

	logger := new(logz.NoOpLogger)

	broker := gochannel.New()

	for _, store := range stores {
		ctx, cancel := context.WithCancel(context.Background())
		eg, egctx := errgroup.WithContext(ctx)

		wg := new(sync.WaitGroup)
		wg.Add(1)

		port, _ := utils.GetFreePort()
		var host = fmt.Sprintf("127.0.0.1:%s", port)

		logFilePath, logFileErr := logz.SuggestedFilePath("ark", "server.log")
		require.NoError(t, logFileErr)

		eg.Go(func() error {
			er := NewSubsystem(host, logFilePath, store, logger, broker).Factory(wg, egctx)()
			require.NoError(t, er)
			t.Log("server shutdown successfully")
			return nil
		})

		wg.Wait()

		client := NewClient(host)
		t.Run("should be able to add a target", func(t *testing.T) {
			_, er := client.AddTarget(ark.RawTarget{
				Name:  "test",
				Type:  docker_image.Type,
				File:  "test",
				Realm: "test",
				Attributes: map[string]interface{}{
					"repo":       "repo",
					"dockerfile": "FROM node",
				},
			})

			require.NoError(t, er)
			// require.True(t, artifact.Cacheable(), "artifact should be cacheable")
		})

		t.Run("should be able to get targets", func(t *testing.T) {
			targetSlice, er := client.GetTargets()
			require.NoError(t, er)
			require.NotEmpty(t, targetSlice)
		})

		t.Run("should be able to connect targets in the graph", func(t *testing.T) {
			_, er := client.ConnectTargets(ark.GraphEdge{
				Src: "test",
				Dst: "test",
			})
			require.NoError(t, er)
		})

		t.Run("should be able to get the graph", func(t *testing.T) {
			_, er := client.GetGraph()
			require.NoError(t, er)
		})

		t.Run("should be able to get all graph edges", func(t *testing.T) {
			_, er := client.GetGraphEdges()
			require.NoError(t, er)
			// require.NotEmpty(t, edges)
		})

		t.Run("should be able to run a target", func(t *testing.T) {
			_, er := client.AddTarget(ark.RawTarget{
				Name:  "test",
				Type:  docker_image.Type,
				File:  "test",
				Realm: "test",
				Attributes: map[string]interface{}{
					"repo":       "repo",
					"dockerfile": "FROM node",
				},
			})

			targetURI := fmt.Sprintf("%s:%s", ".", "test")

			_, err = client.Run(messages.GraphRunnerExecuteCommand{
				TargetKeys: []string{targetURI},
			})
			require.NoError(t, er)
		})

		t.Run("should be able to return an error if target doesn't exist", func(t *testing.T) {

			defer func() {
				t.Log("shutdown signal sent")
				cancel()
			}() // hack and this test suite cannot run in parallel

			targetURI := fmt.Sprintf("%s:%s", ".", "test")

			_, er := client.Run(messages.GraphRunnerExecuteCommand{
				TargetKeys: []string{targetURI},
			})
			require.Error(t, er)

		})
		t.Log("waiting for subsystem to shut down")
		require.NoError(t, eg.Wait())
		t.Log("subsystem exited successfully")
	}

}
