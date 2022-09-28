package group

import (
	"context"

	"github.com/myfintech/ark/src/go/lib/logz"
)

// Action is the executor for targeting a group of Targets
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

// TODO: this method needs a implementation

// Execute runs the action and produces a group.Artifact
func (a Action) Execute(_ context.Context) (err error) {
	return nil
}
