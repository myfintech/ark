package exec

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/myfintech/ark/src/go/lib/log"
	"github.com/myfintech/ark/src/go/lib/utils"
)

// LocalExecOptions for executing a command
type LocalExecOptions struct {
	Command          []string
	Dir              string
	Environment      map[string]string
	Stdin            io.Reader
	Stdout           io.Writer
	Stderr           io.Writer
	InheritParentEnv bool
}

// LocalExecutor executes the task locally
func LocalExecutor(opts LocalExecOptions) *exec.Cmd {
	var cmd *exec.Cmd

	args := utils.ExpandEnvOnArgs(opts.Command)
	cmd = exec.Command(args[0], args[1:]...)
	log.Debugf("Entering working directory %s", opts.Dir)
	log.Debugf("Executing %s", cmd.String())

	cmd.Stdout = opts.Stdout
	cmd.Stderr = opts.Stderr
	cmd.Stdin = opts.Stdin
	cmd.Dir = opts.Dir
	if opts.InheritParentEnv {
		cmd.Env = append(cmd.Env, os.Environ()...)
	}

	if opts.Environment != nil {
		cmd.Env = append(cmd.Env, mapToEnvKVs(opts.Environment)...)
	}

	return cmd
}

func mapToEnvKVs(emap map[string]string) (env []string) {
	for k, v := range emap {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return
}
