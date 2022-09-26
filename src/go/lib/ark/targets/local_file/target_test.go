package local_file

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
		Filename: filepath.Join(testdata, "test.txt"),
		Content:  "This is the content for the test",
		RawTarget: ark.RawTarget{
			Name:  "local_file_test",
			Type:  Type,
			File:  filepath.Join(cwd, "targets_test.go"),
			Realm: cwd,
		},
	}

	err = target.Validate()
	require.NoError(t, err)
}
