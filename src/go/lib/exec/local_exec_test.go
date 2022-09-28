package exec

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalExecutor(t *testing.T) {
	cwd, _ := os.Getwd()
	shell := "/bin/sh"
	cmd := LocalExecutor(LocalExecOptions{
		Command: []string{
			shell,
			"-c",
			"ls -lha .",
		},
		Dir: cwd,
		Environment: map[string]string{
			"GO_TESTING": "test",
		},
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})

	require.NoError(t, cmd.Run())
	require.Equal(t, shell+" -c ls -lha .", cmd.String())
}
