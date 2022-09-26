package container

import (
	"context"
	"io"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/docker/go-connections/nat"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/registry"
)

// Container a generic interface representing container system interactions
type Container interface {
	ParseURL(url string) (reference.NamedTagged, error)
	RepoInfo(ref reference.NamedTagged) (*registry.RepositoryInfo, error)
	Auth(ctx context.Context, ref reference.NamedTagged, cmdName string) (string, types.RequestPrivilegeFunc, error)
	PullImage(ctx context.Context, imageURL string) error
	PushImage(ctx context.Context, imageURL string) error
	RepoGetTags(ctx context.Context, imageURL string) ([]string, error)
	RepoImageExists(ctx context.Context, imageURL string) (bool, error)
	ImageList(ctx context.Context) ([]types.ImageSummary, error)
	ImageExists(ctx context.Context, imageURL string) (bool, error)
	Start(eg *errgroup.Group, ctx context.Context, opts StartOptions) (string, func(), error)
	Wait(ctx context.Context, containerID string, condition WaitCondition) error
	Logs(ctx context.Context, containerID string) (io.ReadCloser, error)
	StreamLogs(stdout, stderr io.Writer, logs io.Reader) error
}

// StartOptions options for the container.Start method
type StartOptions struct {
	AutoRemove    bool
	AttachStdIn   bool
	Privileged    bool
	Image         string
	WorkingDir    string
	ContainerName string
	Binds         []string
	Cmd           []string
	Env           map[string]string
	PortBindings  nat.PortMap
	ExposedPorts  nat.PortSet
	KillTimeout   time.Duration
	InputStream   io.Reader
	OutputStream  io.Writer
}
