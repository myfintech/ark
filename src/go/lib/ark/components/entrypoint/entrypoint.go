package entrypoint

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/myfintech/ark/src/go/lib/utils"

	"github.com/myfintech/ark/src/go/lib/log"

	execute "github.com/myfintech/ark/src/go/lib/exec"
	"github.com/myfintech/ark/src/go/lib/fs"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// entrypoint can be embedded to have forward compatible implementations.
type entrypoint struct {
	SubCmdArgs            []string
	Cmd                   *exec.Cmd
	RestartAfterUnarchive bool
	CommandStop           chan bool
	Ctx                   context.Context
}

// StreamFileChange removes files that have been deleted, and unzips and reads an archive for file changes
func (ep *entrypoint) StreamFileChange(changeStream Sync_StreamFileChangeServer) error {
	log.Info("Connected")
	for {
		cwd, err := os.Getwd()
		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}

		notification, err := changeStream.Recv()
		if err == io.EOF {
			log.Info("Disconnected")
			return nil
		}

		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}

		log.Infof("Received notification for root: %s", notification.Root)

		workspaceRoot := cwd

		if notification.Root != "" {
			workspaceRoot = notification.Root
		}

		for _, f := range notification.Files {
			log.Infof("File change detected: %s", f.RelName)
			if !f.Exists {
				_ = os.Remove(filepath.Join(workspaceRoot, f.RelName))
			}
		}

		if len(notification.Archive) > 0 {
			log.Infof("attempting to decompress %d bytes to %s", len(notification.Archive), workspaceRoot)
			if err = fs.GzipUntar(workspaceRoot, bytes.NewReader(notification.Archive)); err != nil {
				return status.Errorf(codes.Unknown, err.Error())
			}
		}

		for _, action := range notification.Actions {
			command := execute.LocalExecutor(execute.LocalExecOptions{
				Command: action.Command,
				Dir:     action.Workdir,
				// Stdin:            os.Stdin,
				Stdout:           os.Stdout,
				Stderr:           os.Stderr,
				InheritParentEnv: true,
			})
			if runErr := command.Run(); runErr != nil {
				return status.Errorf(codes.Unknown, runErr.Error())
			}
		}

		if ep.RestartAfterUnarchive {
			ep.CommandStop <- true
		}

		if ackErr := changeStream.Send(&FileChangeAck{}); ackErr != nil {
			return status.Errorf(codes.Unknown, ackErr.Error())
		}
	}
}

// Executor uses command line arguments to construct a command from local_exec
func (ep *entrypoint) Executor() *exec.Cmd {
	return execute.LocalExecutor(execute.LocalExecOptions{
		Command: ep.SubCmdArgs,
		// TODO: Review setting the directory
		Dir: "",
		// Stdin:            os.Stdin,
		Stdout:           os.Stdout,
		Stderr:           os.Stderr,
		InheritParentEnv: true,
	})
}

// Watch watches a process and will restart it if it dies pre-maturely
func (ep *entrypoint) Watch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Infof("ep context canceled, exiting watch routine")
			return
		default:
		}

		ep.Cmd = ep.Executor()
		ep.Cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}

		if err := ep.Cmd.Start(); err != nil {
			log.Errorf("failed to start command %s: %v", ep.Cmd.String(), err)
			time.Sleep(2 * time.Second)
			continue
		}

		commandExited := make(chan bool, 1)
		go ep.WaitForStopSignal(ctx, commandExited)
		log.Infof("watching (PID: %d): %s", ep.Cmd.Process.Pid, ep.Cmd.String())

		err := ep.Cmd.Wait()
		log.Infof("process (PID: %d): %s exited with status %d",
			ep.Cmd.Process.Pid, ep.Cmd.String(), ep.Cmd.ProcessState.ExitCode())

		if err != nil {
			log.Errorf("exit error %v", err)
		}

		commandExited <- true

		time.Sleep(2 * time.Second)
		// TODO: We may want to evolve this into an exponential backoff
		// Its not a great experience to log spam a bunch of crash restarts
		// If we don't perform exponential we may want to wait until a new file change is received to attempt the restart instead of restarting a process every two seconds
	}
}

// WaitForStopSignal waits for a cancellation signal and forcefully terminates the command
func (ep *entrypoint) WaitForStopSignal(ctx context.Context, commandExited chan bool) {
	defer log.Info("signal wait completed")
	select {
	case <-ctx.Done():
		log.Info("context canceled, stopping command")
		_ = ep.StopCmd()
		return
	case <-ep.CommandStop:
		log.Info("command stop recieved by channel, stopping command")
		_ = ep.StopCmd()
		return
	case <-commandExited:
		log.Info("command exited, leaving signal wait")
		return
	}
}

// StopCmd sets a timer and sends a kill signal if the timer is not stopped by the interrupt signal
// blocks until it is notified that the command exited
func (ep *entrypoint) StopCmd() error {
	timer := time.AfterFunc(time.Second*10, func() {
		log.Error("process failed to exit within 10 seconds, sending SIGKILL")
		_ = syscall.Kill(-ep.Cmd.Process.Pid, syscall.SIGKILL)
	})

	defer timer.Stop()

	log.Info("sending SIGTERM")
	if err := syscall.Kill(-ep.Cmd.Process.Pid, syscall.SIGTERM); err != nil {
		return err
	}
	return nil
}

// New instantiates a new ark entrypoint server
func New(args []string, ctx context.Context) *entrypoint {
	return &entrypoint{
		SubCmdArgs:            args,
		Cmd:                   nil,
		RestartAfterUnarchive: utils.EnvLookup("ARK_EP_RESTART_MODE", "auto") == "auto",
		CommandStop:           make(chan bool, 1),
		Ctx:                   ctx,
	}
}
