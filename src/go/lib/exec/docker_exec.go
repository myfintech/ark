package exec

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/moby/buildkit/util/appcontext"

	"github.com/docker/go-connections/nat"

	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/container"
)

var appCTX = appcontext.Context()

// DockerExecOptions for executing a command
type DockerExecOptions struct {
	Command          []string
	Binds            []string
	Ports            []string
	ContainerName    string
	KillTimeout      string
	Image            string
	Dir              string
	Environment      map[string]string
	InheritParentEnv bool
	AttachStdIn      bool
	Privileged       bool
	Detach           bool
	Stdin            io.Reader
	Stdout           io.Writer
	Stderr           io.Writer
}

// DockerExecutor executes the task in a docker container
func DockerExecutor(ctx context.Context, opts DockerExecOptions) error {
	if ctx == nil {
		ctx = appCTX
	}

	if opts.KillTimeout == "" {
		opts.KillTimeout = "10s"
	}

	timeoutConversion, err := time.ParseDuration(opts.KillTimeout)
	if err != nil {
		return err
	}

	// TODO: If opts.Detach is true, we need to see if a container is already running and react to that (delete and re-run, maybe, since exec targets aren't cacheable and the 'thing' may have changed?)
	docker, err := container.NewDockerClient(container.DefaultDockerCLIOptions()...)
	if err != nil {
		return errors.Wrap(err, "DockerExecutor#NewDockerClient")
	}

	exists, checkErr := docker.ImageExists(ctx, opts.Image)
	if checkErr != nil {
		return checkErr
	}
	if !exists {
		if err = docker.PullImage(ctx, opts.Image); err != nil {
			return errors.Wrap(err, "#DockerExecutor#PullImage")
		}
	}

	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}

	if opts.Ports != nil {
		for _, port := range opts.Ports {
			// TODO: Move validation into separate function(s) for reusability
			splitPort := strings.Split(port, ":")
			if len(splitPort) != 2 {
				errorMessage := fmt.Sprintf("port binding (%s) was not in the correct format, (hostPort:containerPort)", port)
				return errors.New(errorMessage)
			}
			for _, p := range splitPort {
				intP, convErr := strconv.Atoi(p)
				if convErr != nil {
					errorMessage := fmt.Sprintf("port (%s) is not a number", p)
					return errors.New(errorMessage)
				}
				if intP < 0 || intP > 65535 {
					errorMessage := fmt.Sprintf("port (%s) is not within valid port range (0-65535)", p)
					return errors.New(errorMessage)
				}
			}
			portBindings[nat.Port(splitPort[1])] = []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: splitPort[0]}}
			exposedPorts[nat.Port(splitPort[1])] = struct{}{}
		}
	}

	eg, gctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		containerID, finish, err := docker.Start(eg, gctx, container.StartOptions{
			AutoRemove:    true,
			Privileged:    false,
			WorkingDir:    opts.Dir,
			InputStream:   opts.Stdin,
			Image:         opts.Image,
			Binds:         opts.Binds,
			Cmd:           opts.Command,
			Env:           opts.Environment,
			AttachStdIn:   opts.AttachStdIn,
			ContainerName: opts.ContainerName,
			PortBindings:  portBindings,
			ExposedPorts:  exposedPorts,
			KillTimeout:   timeoutConversion,
		})
		defer finish()
		if err != nil {
			return errors.Wrap(err, "DockerExecutor#StartContainer")
		}

		if opts.Detach {
			return nil
		}

		logs, err := docker.Logs(context.TODO(), containerID)
		if err != nil {
			return errors.Wrap(err, "DockerExecutor#Logs")
		}

		if opts.AttachStdIn {
			if err = docker.StreamLogs(opts.Stdout, os.Stderr, logs); err != nil {
				return errors.Wrap(err, "DockerExecutor#StreamLogs")
			}
		} else if err = docker.StreamLogs(os.Stdout, os.Stderr, logs); err != nil {
			return errors.Wrap(err, "DockerExecutor#StreamLogs")
		}

		if err = docker.Wait(context.TODO(), containerID, container.WaitConditionNotRunning); err != nil {
			return errors.Wrap(err, "DockerExecutor#Wait")
		}
		return nil
	})

	return eg.Wait()
}
