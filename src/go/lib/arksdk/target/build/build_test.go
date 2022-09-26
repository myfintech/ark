package build

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/dag"
)

func TestDockerImageTarget_Build(t *testing.T) {
	cwd, _ := os.Getwd()
	testDataDir := filepath.Join(cwd, "testdata")

	workspace := base.NewWorkspace()
	workspace.RegisteredTargets = base.Targets{
		"build": Target{},
	}

	require.NoError(t, workspace.DetermineRoot(testDataDir))
	require.NoError(t, workspace.InitVaultClient())
	require.NoError(t, workspace.InitDockerClient())

	buildFiles, err := workspace.DecodeBuildFiles()
	require.NoError(t, err, "no error decoding build files")

	require.NoError(t, workspace.LoadTargets(buildFiles), "must load target hcl files into workspace")

	t.Run("Build with secret mount", func(t *testing.T) {
		require.NoError(t, walkByTarget(t, workspace, "test.build.test1"))
	})

	t.Run("Build without secret mount", func(t *testing.T) {
		require.NoError(t, walkByTarget(t, workspace, "test.build.test2"))
	})
}

func findImage(imageURL string) (string, error) {
	ctx := context.Background()
	c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}
	args := filters.NewArgs()
	args.Add("reference", imageURL)
	results, err := c.ImageList(ctx, types.ImageListOptions{
		All:     false,
		Filters: args,
	})
	if err != nil {
		return "", err
	}
	if len(results) < 1 {
		return "", errors.New("no results found")
	}
	summary := results[0]
	if len(summary.RepoTags) < 1 {
		return "", errors.New("no repo tags found")
	}
	return summary.RepoTags[0], nil
}

func walkByTarget(t *testing.T, workspace *base.Workspace, address string) error {
	intendedTarget, err := workspace.TargetLUT.LookupByAddress(address)
	if err != nil {
		return err
	}

	return workspace.GraphWalk(intendedTarget.Address(), func(vertex dag.Vertex) error {
		buildable := vertex.(base.Buildable)
		if preBuildErr := buildable.PreBuild(); preBuildErr != nil {
			return preBuildErr
		}
		if buildErr := buildable.Build(); buildErr != nil {
			return buildErr
		}

		target := buildable.(Target)
		attrs := target.ComputedAttrs()
		imageURL := target.URL(target.Hash())

		if attrs.Output == "" {
			searchResult, lookupErr := findImage(imageURL)
			require.NoError(t, lookupErr)
			require.Equal(t, imageURL, searchResult, "searchResult should be equal to the imageURL")
		}

		require.NotEmpty(t, attrs.Repo, "GetStateAttrs.Repo should not be empty.")
		require.NotEmpty(t, attrs.Dockerfile, "GetStateAttrs.Dockerfile should not be empty.")
		require.NotEmpty(t, attrs.BuildArgs, "GetStateAttrs.BuildArgs should not be empty.")
		require.NotEmpty(t, attrs.Tags, "GetStateAttrs.Tags should not be empty.")
		return nil
	})
}
