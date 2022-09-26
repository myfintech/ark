package nix

import (
	"context"
	"os"
	"os/exec"

	"github.com/myfintech/ark/src/go/lib/logz"
)

// Action is the executor for installing a nix package
type Action struct {
	Artifact *Artifact
	Target   *Target
	Logger   logz.FieldLogger
}

var _ logz.Injector = &Action{}

// UseLogger injects a logger into the target's action
func (a *Action) UseLogger(logger logz.FieldLogger) {
	a.Logger = logger
}

// Execute runs the action and produces a nix.Artifact
func (a Action) Execute(_ context.Context) (err error) {
	for _, nixPkg := range a.Target.Packages {
		cmd := exec.Command("nix-env", "--quiet", "-iA", nixPkg)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}
