package group

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

	target := Target{
		RawTarget: ark.RawTarget{
			Name:  "group_test",
			Type:  Type,
			File:  filepath.Join(cwd, "targets_test.go"),
			Realm: cwd,
		},
	}

	err = target.Validate()
	require.NoError(t, err)
}
