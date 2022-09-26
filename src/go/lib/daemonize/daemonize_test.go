package daemonize

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStart(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	testdata := filepath.Join(cwd, "testdata")
	err = os.MkdirAll(testdata, 0755)
	require.NoError(t, err)

	pidfile := filepath.Join(testdata, "pid")

	pid, err := Fork("sleep", "3600")
	require.NoError(t, err)

	err = ioutil.WriteFile(pidfile, []byte(strconv.Itoa(pid)), 0755)
	require.NoError(t, err)

	err = Stop(pidfile)
	require.NoError(t, err)

	proc := NewProc("sleep", []string{"3606"}, pidfile)
	err = proc.Init()
	require.NoError(t, err)

	err = proc.Init()
	require.ErrorIs(t, err, ErrAlreadyRunning)

	err = proc.Stop()
	require.NoError(t, err)
}
