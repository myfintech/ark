package exec

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"golang.org/x/sync/errgroup"
)

func TestDockerExecutor(t *testing.T) {
	t.Run("Runs to completion", func(t *testing.T) {
		testEnv := map[string]string{"TEST": "value"}

		err := DockerExecutor(nil, DockerExecOptions{
			Command: []string{
				"bash",
				"-c",
				"apt update -y && apt upgrade -y && [[ $TEST == 'value' ]] && echo 'success' || echo 'fail'",
			},
			Dir:              "/",
			Image:            "ubuntu:latest",
			Environment:      testEnv,
			Stdin:            os.Stdin,
			Stdout:           os.Stdout,
			Stderr:           os.Stderr,
			InheritParentEnv: false,
		})

		require.NoError(t, err)
	})
	t.Run("Cancel the context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		eg, gctx := errgroup.WithContext(ctx)
		eg.Go(func() error {
			return DockerExecutor(gctx, DockerExecOptions{
				Command: []string{
					"sh",
					"-c",
					"sleep 30",
				},
				Dir:              "/",
				Image:            "alpine:latest",
				Environment:      nil,
				Stdin:            os.Stdin,
				Stdout:           os.Stdout,
				Stderr:           os.Stderr,
				InheritParentEnv: false,
				KillTimeout:      "3s",
			})
		})
		eg.Go(func() error {
			time.Sleep(5 * time.Second)
			cancel()
			t.Log("called context cancel func")
			return nil
		})
		require.Error(t, eg.Wait())
	})
}
