package local_file

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/pkg/errors"
)

// Action is the executor for writing a local file to disk
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

// Execute runs the action and produces a local_file.Artifact
func (a Action) Execute(_ context.Context) error {
	fileBytes := []byte(a.Target.Content)

	if err := os.MkdirAll(filepath.Dir(a.Artifact.RenderedFilePath), 0755); err != nil {
		return errors.Wrap(err, "unable to create directory structure")
	}

	if err := ioutil.WriteFile(a.Artifact.RenderedFilePath, fileBytes, 0755); err != nil {
		return errors.Wrap(err, "unable to write file")
	}

	return nil
}
