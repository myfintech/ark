package deploy

import (
	"encoding/hex"
	"hash"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/kube/portbinder"
)

// Type is the string value of the RawTarget type
const Type = "deploy"

// Step defines the required fields to execute one or more commands when a changed file matches a given pattern
type Step struct {
	Command  []string `mapstructure:"manifest" json:"manifest"`
	WorkDir  string   `mapstructure:"workDir"  json:"workDir"`
	Patterns []string `mapstructure:"patterns" json:"patterns"`
}

// Target expresses an intention to construct a deployable manifest
type Target struct {
	ark.RawTarget       `mapstructure:",squash"`
	Manifest            string              `mapstructure:"manifest"            json:"manifest"`
	PortForward         portbinder.PortMap  `mapstructure:"portForward"         json:"portForward"` // TODO: There's no input validation for this in this code
	LiveSyncEnabled     bool                `mapstructure:"liveSyncEnabled"     json:"liveSyncEnabled"`
	LiveSyncRestartMode string              `mapstructure:"liveSyncRestartMode" json:"liveSyncRestartMode"` // TODO: There's no input validation for this code or setting of a default
	LiveSyncOnStep      []Step              `mapstructure:"liveSyncOnStep"      json:"liveSyncOnStep"`
	Env                 []map[string]string `mapstructure:"env"                 json:"env"`
}

// Produce should produce a deterministic manifest artifact
func (t Target) Produce(checksum hash.Hash) (ark.Artifact, error) {
	artifact := &Artifact{
		RawArtifact: ark.RawArtifact{
			Key:        t.Key(),
			Type:       t.Type,
			Hash:       hex.EncodeToString(checksum.Sum(nil)),
			Attributes: t.Attributes,
			DependsOn:  t.DependsOn,
		},
	}
	return artifact, nil
}

// Validate checks if the RawTarget fields are valid
func (t *Target) Validate() error {
	if err := t.RawTarget.Validate(); err != nil {
		return err
	}
	return validation.ValidateStruct(t,
		validation.Field(&t.Manifest, validation.Required),
	)
}
