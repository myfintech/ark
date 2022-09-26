package local_file

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/myfintech/ark/src/go/lib/ark"

	"github.com/stretchr/testify/require"
)

func TestAction(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	testdata := filepath.Join(cwd, "testdata")

	target := &Target{
		Filename: filepath.Join(testdata, "test.txt"),
		Content:  "This is the content for the test",
		RawTarget: ark.RawTarget{
			Name:  "local_file_test",
			Type:  Type,
			Realm: cwd,
			File:  filepath.Join(cwd, "targets_test.go"),
		},
	}

	err = target.Validate()
	require.NoError(t, err)

	hash, err := target.Checksum()
	require.NoError(t, err)

	artifact, err := target.Produce(hash)
	require.NoError(t, err)

	action := &Action{
		Target:   target,
		Artifact: artifact.(*Artifact),
	}

	err = action.Execute(context.Background())
	require.NoError(t, err)
}
