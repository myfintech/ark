package docker_image

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/myfintech/ark/src/go/lib/ark"

	"github.com/stretchr/testify/require"
)

func TestArtifact(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	testdata := filepath.Join(cwd, "testdata")

	target := &Target{
		Repo:       "gcr.io/managed-infrastructure/ark/test",
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

	hash, err := target.Checksum()
	require.NoError(t, err)

	artifact, err := target.Produce(hash)
	require.NoError(t, err)

	image := artifact.(*Artifact)
	image.Client = client
	require.True(t, image.Cacheable())

	action := &Action{
		Client:   client,
		Target:   target,
		Artifact: image,
	}

	err = action.Execute(ctx)
	require.NoError(t, err)

	locallyCached, err := image.LocallyCached(ctx)
	require.NoError(t, err)
	require.True(t, locallyCached)

	err = image.Push(ctx)
	require.NoError(t, err)

	remotelyCached, err := image.RemotelyCached(ctx)
	require.NoError(t, err)
	require.True(t, remotelyCached)
}
