package daemonize

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/juju/fslock"
	"github.com/pkg/errors"
)

// STATE the applications execution state
type STATE int

const (
	// STATE_UNKNOWN the process is in an unknown state (possible due to an error)
	STATE_UNKNOWN STATE = iota

	// STATE_RUNNING the process is running
	STATE_RUNNING

	// STATE_STOPPED the process isn't currently running
	STATE_STOPPED
)

var (
	// ErrPidNotFound raised when a pid cannot be by querying the OS
	ErrPidNotFound = errors.New("pid not found")

	// ErrPIDInvalid raised when an invalid pid is parsed from a pidfile
	ErrPIDInvalid = errors.New("pid wasn't valid")

	// ErrFailedToAcquirePidLock raised when a lock acquisition times out or fails due to contention
	ErrFailedToAcquirePidLock = errors.New("failed to acquire lock on pidfile")

	// ErrFailedToWritePidFile raised when the pid cannot be written to the pidfile
	ErrFailedToWritePidFile = errors.New("failed to write pidfile")

	// ErrFailedToStartProc we failed to call os.Exec
	ErrFailedToStartProc = errors.New("failed to start process")

	// ErrFailedToDetach we started the process but failed to orphan it
	ErrFailedToDetach = errors.New("failed to detach from process")

	// ErrAlreadyRunning raised by Proc.Init when a pid is active
	ErrAlreadyRunning = errors.New("cannot init process pid is already running")

	// ErrTimeout is raised when we fail to stop a process
	ErrTimeout = errors.New("an operation timed out")
)

// Proc a struct that houses data for managing a process
type Proc struct {
	command   string
	arguments []string
	pidfile   string
	lock      *fslock.Lock
}

func (p *Proc) acquirePidLock() error {
	p.lock = fslock.New(p.pidfile)

	err := p.lock.LockWithTimeout(time.Second * 5)
	if err == nil {
		return nil
	}

	switch err {
	case fslock.ErrTimeout:
		return errors.Wrapf(ErrFailedToAcquirePidLock, "timed out on %s", p.pidfile)
	default:
		return errors.Wrapf(ErrFailedToAcquirePidLock, "%v", err)
	}
}

// Init attempts to validate if a process is already running at the given pidfile
// If the process is already running no action is taken
// If the process isn't running we attempt to acquire a lock pidfile
// We then start the process and store its pid in the pidfile and release the lock
func (p *Proc) Init() error {
	if pid, err := LoadPidFile(p.pidfile); err == nil {
		if _, err = FindProcess(pid); err == nil {
			return errors.Wrapf(ErrAlreadyRunning, "pid %d", pid)
		}
	}

	if err := os.MkdirAll(filepath.Dir(p.pidfile), 0755); err != nil {
		return errors.Wrap(err, "failed to create pid file directory")
	}

	if err := p.acquirePidLock(); err != nil {
		return err
	}

	defer p.releasePidLock()

	pid, err := Fork(p.command, p.arguments...)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(p.pidfile, []byte(strconv.Itoa(pid)), 0755)
	if err != nil {
		return errors.Wrapf(ErrFailedToWritePidFile, "%v", err)
	}

	return nil
}

// Stop stops the current proc
func (p *Proc) Stop() error {
	return Stop(p.pidfile)
}

func (p *Proc) releasePidLock() {
	if p.lock == nil {
		return
	}
	if err := p.lock.Unlock(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to relrease lock on %s %v \n", p.pidfile, err)
	}
}

// Status returns the STATE of the running process
func (p *Proc) Status() (STATE, error) {
	pid, err := LoadPidFile(p.pidfile)
	if err != nil {
		return STATE_UNKNOWN, err
	}

	_, err = FindProcess(pid)
	if errors.Is(err, ErrPidNotFound) {
		return STATE_STOPPED, nil
	}

	if err != nil {
		return STATE_UNKNOWN, err
	}

	return STATE_RUNNING, nil
}

// NewProc creates a new instance of a Proc init system
func NewProc(command string, arguments []string, pidfile string) *Proc {
	return &Proc{command: command, arguments: arguments, pidfile: pidfile}
}

// Fork allows a program to be started and orphaned
// This process will exit without cleaning up its child process
// It is highly recommended managing the PID returned from this function
func Fork(program string, args ...string) (int, error) {
	cmd := exec.Command(program, args...)
	if err := cmd.Start(); err != nil {
		return 0, errors.Wrapf(ErrFailedToStartProc, "%v", err)
	}

	pid := cmd.Process.Pid

	if err := cmd.Process.Release(); err != nil {
		return 0, errors.Wrapf(ErrFailedToDetach, "%v", err)
	}

	return pid, nil
}

// Stop accepts the location of a pid file and attempts to clean up the running process
// Will block until the process exits and timeout after 10 seconds
func Stop(pidfile string) error {
	pid, err := LoadPidFile(pidfile)
	if err != nil {
		return err
	}

	proc, err := FindProcess(pid)
	if err != nil {
		return err
	}

	if err = proc.Kill(); err != nil {
		return errors.Wrapf(err, "failed to kil pid %d", pid)
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		_, _ = proc.Wait()
	}()

	select {
	case <-done:
		return nil
	case <-time.After(time.Second * 10):
		return errors.Wrapf(ErrTimeout, "failed to stop process %s", pidfile)
	}
}

// FindProcess attempts to locate a process by its pid
// returns a wrapped ErrPidNotFound if the process cannot be signaled
func FindProcess(pid int) (*os.Process, error) {
	// always succeeds on UNIX systems
	proc, _ := os.FindProcess(pid)

	// If  sig  is 0, then no signal is sent, but error checking is still performed.
	// This can be used to check for the existence of a process ID  or process group ID.
	err := proc.Signal(syscall.Signal(0))
	if err != nil {
		return proc, errors.Wrapf(ErrPidNotFound, "failed to locate pid %d %v", pid, err)
	}
	return proc, nil
}

// LoadPidFile opens a pidfile and parses its value as an integer
func LoadPidFile(pidfile string) (int, error) {
	data, err := ioutil.ReadFile(pidfile)
	if err != nil {
		return 0, errors.Wrap(err, "failed to read pid file")
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return pid, errors.Wrapf(ErrPIDInvalid, "invalid pid found in %s", pidfile)
	}

	return pid, nil
}
