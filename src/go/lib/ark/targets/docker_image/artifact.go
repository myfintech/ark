package docker_image

import (
	"context"

	"github.com/myfintech/ark/src/go/lib/ark"

	"github.com/myfintech/ark/src/go/lib/container"
)

// Artifact is the result of a successful actions.BuildDockerImage
type Artifact struct {
	ark.RawArtifact `mapstructure:",squash"`
	URL             string           `json:"url" mapstructure:"url"`
	Client          container.Docker `json:"-" mapstructure:"-"`
}

// UseDockerClient sets this Artifact container.Docker client
func (a *Artifact) UseDockerClient(client container.Docker) {
	a.Client = client
}

// Cacheable always returns true because docker image may be stored remotely and locally
func (a Artifact) Cacheable() bool {
	return true
}

// RemotelyCached determines if the image exists in the remote repository by its URL
func (a Artifact) RemotelyCached(ctx context.Context) (bool, error) {
	return a.Client.RepoImageExists(ctx, a.URL)
}

// LocallyCached determines if the image exists in the local docker registry by its URL
func (a Artifact) LocallyCached(ctx context.Context) (bool, error) {
	return a.Client.ImageExists(ctx, a.URL)
}

// Push attempts to upload the docker image to the remote registry by its URL
func (a Artifact) Push(ctx context.Context) error {
	return a.Client.PushImage(ctx, a.URL)
}

// Pull attempts to pull the docker image from the remote registry by its URL
func (a Artifact) Pull(ctx context.Context) error {
	return a.Client.PullImage(ctx, a.URL)
}
