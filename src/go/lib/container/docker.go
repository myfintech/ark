package container

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/moby/buildkit/session/secrets/secretsprovider"

	"github.com/myfintech/ark/src/go/lib/log"

	"github.com/containerd/console"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/docker/registry"
	"github.com/docker/go-connections/tlsconfig"
	controlapi "github.com/moby/buildkit/api/services/control"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/session/sshforward/sshprovider"
	"github.com/moby/buildkit/session/upload/uploadprovider"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/myfintech/ark/src/go/lib/utils"
)

var ctxLog = log.WithFields(log.Fields{
	"prefix": "container",
})

// WaitCondition provides specificity to the type of string recognized as a container wait condition
type WaitCondition string

const (
	// WaitConditionNotRunning defines the 'not-running' condition to check while waiting
	WaitConditionNotRunning WaitCondition = "not-running"

	// Comments added for public consts, but these two aren't used ... yet ...
	// WaitConditionNextExit defines the 'next-exit' condition to check while waiting
	// WaitConditionNextExit   WaitCondition = "next-exit"
	// WaitConditionRemoved defines the 'removed' condition to check while waiting
	// WaitConditionRemoved    WaitCondition = "removed"
)

// Docker an implementation of the container interface
type Docker struct {
	cli          *command.DockerCli
	OutputWriter io.Writer
}

// ParseURL returns a docker repository reference from a given url string
func (d *Docker) ParseURL(url string) (reference.NamedTagged, error) {
	if !strings.Contains(url, ":") {
		url = fmt.Sprintf("%s:latest", url)
	}
	ref, err := reference.ParseNormalizedNamed(url)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing %s", url)
	}

	nt, ok := ref.(reference.NamedTagged)
	if !ok {
		return nil, errors.Errorf("could not parse ref %s as reference.NamedTagged", ref)
	}
	return nt, nil
}

// RepoInfo returns the repository information for a given image reference
func (d *Docker) RepoInfo(ref reference.NamedTagged) (*registry.RepositoryInfo, error) {
	return registry.ParseRepositoryInfo(ref)
}

// Auth returns credentials for operations that require authentication
func (d *Docker) Auth(ctx context.Context, ref reference.NamedTagged, cmdName string) (string, types.RequestPrivilegeFunc, error) {
	repoInfo, err := d.RepoInfo(ref)
	if err != nil {
		return "", nil, errors.Wrap(err, "Docker.Auth#ParseRepositoryInfo")
	}

	configKey := repoInfo.Index.Name
	if repoInfo.Index.Official {
		configKey = command.ElectAuthServer(ctx, d.cli)
	}
	authConfig, err := d.cli.ConfigFile().GetAuthConfig(configKey)
	if err != nil {
		return "", nil, err
	}
	requestPrivilege := command.RegistryAuthenticationPrivilegedFunc(d.cli, repoInfo.Index, cmdName)

	encodedAuth, err := command.EncodeAuthToBase64(types.AuthConfig(authConfig))
	if err != nil {
		return "", nil, errors.Wrap(err, "Docker.Auth#EncodeAuthToBase64")
	}

	return encodedAuth, requestPrivilege, nil
}

// PullImage pulls a docker image from the given url
func (d *Docker) PullImage(ctx context.Context, imageURL string) error {
	ref, err := d.ParseURL(imageURL)
	if err != nil {
		return errors.Wrap(err, "Docker.PullImage#ParseURL")
	}

	encodedAuth, requestPrivilege, err := d.Auth(ctx, ref, "pull")
	if err != nil {
		return err
	}

	pullStream, err := d.cli.Client().ImagePull(ctx, imageURL, types.ImagePullOptions{
		RegistryAuth:  encodedAuth,
		PrivilegeFunc: requestPrivilege,
	})
	if err != nil {
		return errors.Wrap(err, "Docker.PullImage#ImagePull")
	}

	// defer pullStream.Close()

	termFd, isTerm := term.GetFdInfo(d.OutputWriter)
	return jsonmessage.DisplayJSONMessagesStream(pullStream, d.OutputWriter, termFd, isTerm, nil)
}

// PushImage pushes a docker image to the given url
func (d *Docker) PushImage(ctx context.Context, imageURL string) error {
	ref, err := d.ParseURL(imageURL)
	if err != nil {
		return errors.Wrap(err, "Docker.PushImage#ParseURL")
	}

	encodedAuth, requestPrivilege, err := d.Auth(ctx, ref, "push")
	if err != nil {
		return err
	}

	pullStream, err := d.cli.Client().ImagePush(ctx, imageURL, types.ImagePushOptions{
		RegistryAuth:  encodedAuth,
		PrivilegeFunc: requestPrivilege,
	})
	if err != nil {
		return errors.Wrap(err, "Docker.PushImage#ImagePull")
	}

	defer func() {
		_ = pullStream.Close()
	}()

	termFd, isTerm := term.GetFdInfo(d.OutputWriter)
	return jsonmessage.DisplayJSONMessagesStream(pullStream, d.OutputWriter, termFd, isTerm, nil)
}

// RepoGetTags searches a remote repository and lists all tags for that image
func (d *Docker) RepoGetTags(ctx context.Context, imageURL string) ([]string, error) {
	ref, err := d.ParseURL(imageURL)
	if err != nil {
		return nil, errors.Wrap(err, "Docker.RepoGetTags#ParseURL")
	}

	tags, err := d.cli.RegistryClient(false).GetTags(ctx, ref)
	if err != nil {
		return tags, errors.Wrap(err, "Docker.RepoGetTags#GetTags")
	}
	return tags, nil
}

// RepoImageExists searches for the existence of an image in a remote repo
func (d *Docker) RepoImageExists(ctx context.Context, imageURL string) (bool, error) {
	if !strings.Contains(imageURL, ":") {
		imageURL = fmt.Sprintf("%s:latest", imageURL)
	}
	tags, err := d.RepoGetTags(ctx, imageURL)
	if err != nil {
		return false, errors.Wrap(err, "Docker.RepoImageExists#RepoGetTags")
	}
	for _, tag := range tags {
		if strings.HasSuffix(imageURL, tag) {
			return true, nil
		}
	}
	return false, nil
}

// ImageList returns the local list of images
func (d *Docker) ImageList(ctx context.Context) ([]types.ImageSummary, error) {
	results, err := d.cli.Client().ImageList(ctx, types.ImageListOptions{
		All: true,
	})
	if err != nil {
		return results, errors.Wrap(err, "Docker.ImageList#ImageList")
	}
	return results, nil
}

// ImageExists returns true if an image exists locally
func (d *Docker) ImageExists(ctx context.Context, imageURL string) (bool, error) {
	if !strings.Contains(imageURL, ":") {
		imageURL = fmt.Sprintf("%s:latest", imageURL)
	}
	images, err := d.ImageList(ctx)
	if err != nil {
		return false, errors.Wrap(err, "Docker.ImageExists#ImageList")
	}
	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == imageURL {
				return true, nil
			}
		}
	}
	return false, nil
}

// Start starts a docker container using the given image reference.
// The finish callback must be executed for this function's error group
// to complete, exiting the run
//
// Example:
//
//	containerID, finish, err := docker.Start()
//	defer finish()
func (d *Docker) Start(eg *errgroup.Group, ctx context.Context, opts StartOptions) (containerID string, finish func(), err error) {
	resp, err := d.cli.Client().ContainerCreate(ctx, &dockerContainer.Config{
		Cmd:          opts.Cmd,
		Image:        opts.Image,
		WorkingDir:   opts.WorkingDir,
		ExposedPorts: opts.ExposedPorts,
		Env:          utils.MapToEnvStringSlice(opts.Env),
		AttachStdin:  opts.AttachStdIn,
		OpenStdin:    opts.AttachStdIn,
		StdinOnce:    opts.AttachStdIn,
	}, &dockerContainer.HostConfig{
		AutoRemove:   opts.AutoRemove,
		Privileged:   opts.Privileged,
		Binds:        opts.Binds,
		PortBindings: opts.PortBindings,
		NetworkMode:  dockerContainer.NetworkMode(opts.NetworkMode),
	}, nil, opts.ContainerName)
	if err != nil {
		return "", func() {}, errors.Wrap(err, "Docker.Start#ContainerCreate")
	}

	if opts.AttachStdIn {
		waiter, attachErr := d.cli.Client().ContainerAttach(ctx, resp.ID, types.ContainerAttachOptions{
			Stream: true,
			Stdin:  true,
			Stdout: false,
			Stderr: false,
		})
		if attachErr != nil {
			return "", func() {}, errors.Wrap(attachErr, "Docker.Start#ContainerAttach")
		}
		inputBytes, readErr := ioutil.ReadAll(opts.InputStream)
		if readErr != nil {
			return "", func() {}, errors.Wrap(readErr, "Docker.Start#IOUtil.ReadAll")
		}
		_, err = waiter.Conn.Write(inputBytes)
		waiter.Close()
		if err != nil {
			return "", func() {}, errors.Wrap(err, "Docker.Start#ContainerAttach.Conn.Write")
		}
	}

	if err = d.cli.Client().ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return resp.ID, func() {}, errors.Wrap(err, "Docker.Start#ContainerStart")
	}

	done := make(chan struct{})
	eg.Go(func() error {
		select {
		case <-ctx.Done():
			ctxLog.Infof("shutdown signal received; killing container %s", resp.ID)
			return d.cli.Client().ContainerStop(context.TODO(), resp.ID, &opts.KillTimeout)
		case <-done:
		}
		return nil
	})

	return resp.ID, func() {
		close(done)
	}, nil
}

// Build starts a docker build
func (d *Docker) Build(eg *errgroup.Group, ctx context.Context, dockerContext io.Reader, opts types.ImageBuildOptions) error {
	resp, err := d.cli.Client().ImageBuild(ctx, dockerContext, opts)
	if err != nil {
		return errors.Wrap(err, "Docker.Build#Submit")
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	done := make(chan struct{})
	defer close(done)
	eg.Go(func() error {
		select {
		case <-ctx.Done():
			return d.cli.Client().BuildCancel(context.TODO(), opts.BuildID)
		case <-done:
		}
		return nil
	})

	t := newTracer()
	termFd, isTerm := term.GetFdInfo(d.OutputWriter)
	defer close(t.displayCh)

	eg.Go(func() error {
		var c console.Console
		return progressui.DisplaySolveStatus(context.TODO(), "", c, d.OutputWriter, t.displayCh)
	})

	return jsonmessage.DisplayJSONMessagesStream(resp.Body, d.OutputWriter, termFd, isTerm, newAuxWriter(t, nil))
}

// Wait for a container condition to change
func (d *Docker) Wait(ctx context.Context, containerID string, condition WaitCondition) error {
	statusCh, errCh := d.cli.Client().ContainerWait(ctx, containerID, dockerContainer.WaitCondition(condition))
	select {
	case err := <-errCh:
		if err != nil {
			return errors.Wrap(err, "Docker.Wait#ContainerWait")
		}
	case status := <-statusCh:
		if status.StatusCode != 0 {
			errMsg := "Unknown"
			if status.Error != nil {
				errMsg = status.Error.Message
			}
			return errors.Errorf("the container exited unsuccessfully with error: %s, error code: %d", errMsg, status.StatusCode)
		}
	}
	return nil
}

// Logs a log stream for the given containerID
func (d *Docker) Logs(ctx context.Context, containerID string) (io.ReadCloser, error) {
	return d.cli.Client().ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
}

// StreamLogs streams logs to the given destinations
func (d *Docker) StreamLogs(stdout, stderr io.Writer, logs io.Reader) error {
	_, err := stdcopy.StdCopy(stdout, stderr, logs)
	if err != nil {
		return errors.Wrap(err, "Docker.StreamLogs#StdCopy")
	}
	return nil
}

// StartBuildkitSession
// func (d *Docker) StartBuildkitSession(ctx context.Context, key string, sshSpecs []string, secretSpecs []string) (*session.Session, error) {
func (d *Docker) StartBuildkitSession(eg *errgroup.Group, ctx context.Context, secretSpecs []secretsprovider.FileSource, key string) (*session.Session, error) {
	bkSession, err := session.NewSession(ctx, "ark", key)
	if err != nil {
		return nil, err

	}

	bkSession.Allow(authprovider.NewDockerAuthProvider(d.OutputWriter))
	bkSession.Allow(uploadprovider.New())

	if sshAuthSock := os.Getenv("SSH_AUTH_SOCK"); sshAuthSock != "" {
		sshAgent, sshConfErr := sshprovider.NewSSHAgentProvider([]sshprovider.AgentConfig{
			{Paths: []string{sshAuthSock}},
		})
		if sshConfErr != nil {
			return bkSession, sshConfErr
		}
		bkSession.Allow(sshAgent)
	}

	if len(secretSpecs) > 0 {
		store, secretStoreErr := secretsprovider.NewFileStore(secretSpecs)
		if secretStoreErr != nil {
			return nil, errors.Wrapf(secretStoreErr, "could not create a new secret store: %v", secretSpecs)
		}

		provider := secretsprovider.NewSecretProvider(store)
		bkSession.Allow(provider)
	}
	//
	// if len(sshSpecs) > 0 {
	// 	sshp, err := buildkit.ParseSSHSpecs(sshSpecs)
	// 	if err != nil {
	// 		return nil, errors.Wrapf(err, "could not parse ssh: %v", sshSpecs)
	// 	}
	// 	bkSession.Allow(sshp)
	// }

	eg.Go(func() error {
		defer func() {
			_ = bkSession.Close()
		}()

		return bkSession.Run(ctx, func(ctx context.Context, proto string, meta map[string][]string) (net.Conn, error) {
			return d.cli.Client().DialHijack(ctx, "/session", proto, meta)
		})
	})
	return bkSession, nil
}

// NewDockerClient returns an instance of the Docker container interface
func NewDockerClient(cliOptions ...command.DockerCliOption) (*Docker, error) {
	cli, err := command.NewDockerCli(cliOptions...)
	if err != nil {
		return nil, errors.Wrap(err, "NewDockerClient#NewDockerCli")
	}

	opts := flags.NewClientOptions()
	if dockerCertPath := os.Getenv("DOCKER_CERT_PATH"); dockerCertPath != "" {
		opts.Common.TLSOptions = &tlsconfig.Options{
			CAFile:             filepath.Join(dockerCertPath, "ca.pem"),
			CertFile:           filepath.Join(dockerCertPath, "cert.pem"),
			KeyFile:            filepath.Join(dockerCertPath, "key.pem"),
			InsecureSkipVerify: os.Getenv("DOCKER_TLS_VERIFY") == "",
		}
		opts.Common.TLSVerify = os.Getenv("DOCKER_TLS_VERIFY") == "1"
	}

	err = cli.Initialize(opts)
	if err != nil {
		return nil, errors.Wrap(err, "NewDockerClient#InitializeCLI")
	}
	return &Docker{
		cli:          cli,
		OutputWriter: os.Stderr,
	}, nil
}

// DefaultDockerCLIOptions returns default docker cli options which attach cli streams to stdio
func DefaultDockerCLIOptions() []command.DockerCliOption {
	return []command.DockerCliOption{
		command.WithOutputStream(os.Stdout),
		command.WithErrorStream(os.Stderr),
		command.WithInputStream(os.Stdin),
		command.WithContentTrust(true),
	}
}

type tracer struct {
	displayCh chan *client.SolveStatus
}

func newTracer() *tracer {
	return &tracer{
		displayCh: make(chan *client.SolveStatus),
	}
}

func (t *tracer) write(msg jsonmessage.JSONMessage) {
	var resp controlapi.StatusResponse

	if msg.ID != "moby.buildkit.trace" {
		return
	}

	var dt []byte
	// ignoring all messages that are not understood
	if err := json.Unmarshal(*msg.Aux, &dt); err != nil {
		return
	}
	if err := (&resp).Unmarshal(dt); err != nil {
		return
	}

	s := client.SolveStatus{}
	for _, v := range resp.Vertexes {
		s.Vertexes = append(s.Vertexes, &client.Vertex{
			Digest:    v.Digest,
			Inputs:    v.Inputs,
			Name:      v.Name,
			Started:   v.Started,
			Completed: v.Completed,
			Error:     v.Error,
			Cached:    v.Cached,
		})
	}
	for _, v := range resp.Statuses {
		s.Statuses = append(s.Statuses, &client.VertexStatus{
			ID:        v.ID,
			Vertex:    v.Vertex,
			Name:      v.Name,
			Total:     v.Total,
			Current:   v.Current,
			Timestamp: v.Timestamp,
			Started:   v.Started,
			Completed: v.Completed,
		})
	}
	for _, v := range resp.Logs {
		s.Logs = append(s.Logs, &client.VertexLog{
			Vertex:    v.Vertex,
			Stream:    int(v.Stream),
			Data:      v.Msg,
			Timestamp: v.Timestamp,
		})
	}

	t.displayCh <- &s
}

func newAuxWriter(t *tracer, errCallback func(error)) func(jsonmessage.JSONMessage) {
	return func(msg jsonmessage.JSONMessage) {
		if msg.ID == "moby.image.id" {
			var result types.BuildResult
			if err := json.Unmarshal(*msg.Aux, &result); err != nil {
				if errCallback != nil {
					errCallback(err)
				}
			}
			return
		}
		t.write(msg)
	}
}
