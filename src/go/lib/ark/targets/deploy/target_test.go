package deploy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/myfintech/ark/src/go/lib/ark"

	"github.com/stretchr/testify/require"
)

func TestTarget(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	testdata := filepath.Join(cwd, "testdata")

	target := Target{
		Manifest:            filepath.Join(testdata, "deploy_test/before.yaml"),
		PortForward:         nil,
		LiveSyncEnabled:     false,
		LiveSyncRestartMode: "",
		LiveSyncOnStep:      nil,
		Env:                 nil,
		RawTarget: ark.RawTarget{
			Name:  "deploy_test",
			Type:  Type,
			File:  filepath.Join(cwd, "targets_test.go"),
			Realm: cwd,
		},
	}

	err = target.Validate()
	require.NoError(t, err)
}
