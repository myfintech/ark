package docker_image

import (
	"crypto/sha256"
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
		Repo:       "gcr.io",
		Dockerfile: "FROM node",
		RawTarget: ark.RawTarget{
			Name:  "example",
			Realm: cwd,
			Type:  Type,
			File:  filepath.Join(cwd, "targets_test.go"),
			SourceFiles: []string{
				filepath.Join(testdata, "01_dont_change_me.txt"),
				filepath.Join(testdata, "02_dont_change_me.txt"),
			},
		},
	}

	err = target.Validate()
	require.NoError(t, err)

	artifact, err := target.Produce(sha256.New())
	require.NoError(t, err)

	image := artifact.(*Artifact)
	require.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", image.Hash)
	require.Equal(t, "gcr.io:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", image.URL)
}
