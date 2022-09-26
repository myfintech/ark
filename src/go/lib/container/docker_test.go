package container

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocker(t *testing.T) {
	docker, cErr := NewDockerClient(DefaultDockerCLIOptions()...)
	require.NoError(t, cErr)
	ctx := context.Background()

	t.Run("should be able to pull an image", func(t *testing.T) {
		require.NoError(t, docker.PullImage(ctx, "node:latest"))
	})

	t.Run("should be able to pull an image without a tag", func(t *testing.T) {
		require.NoError(t, docker.PullImage(ctx, "node"))
	})

	t.Run("should be able to list local images", func(t *testing.T) {
		images, err := docker.ImageList(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, images)
	})

	t.Run("should be able to check the existence of a local image", func(t *testing.T) {
		exists, err := docker.ImageExists(ctx, "node:latest")
		require.NoError(t, err)
		require.Equal(t, true, exists)
	})

	t.Run("should be able to check the existence of a local image without a tag", func(t *testing.T) {
		exists, err := docker.ImageExists(ctx, "node")
		require.NoError(t, err)
		require.Equal(t, true, exists)
	})

	t.Run("can search for an image by its URL", func(t *testing.T) {
		exists, err := docker.RepoImageExists(ctx, "node:latest")
		require.NoError(t, err)
		require.Equal(t, true, exists)
	})

	t.Run("can search for an image by its URL without a tag", func(t *testing.T) {
		exists, err := docker.RepoImageExists(ctx, "node")
		require.NoError(t, err)
		require.Equal(t, true, exists)
	})

	t.Run("should be able to pull an image from a private registry", func(t *testing.T) {
		require.NoError(t, docker.PullImage(ctx, "gcr.io/managed-infrastructure/mantl/node-10:latest"))
	})
}
