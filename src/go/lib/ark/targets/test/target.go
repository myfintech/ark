package test

import (
	"encoding/hex"
	"hash"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/myfintech/ark/src/go/lib/ark"
)

// Type is the string value of the Target type
const Type = "test"

// Target expresses the intention to implement a Test target
type Target struct {
	ark.RawTarget    `mapstructure:",squash"`
	Command          []string          `json:"command" mapstructure:"command"`
	Args             []string          `json:"args" mapstructure:"args"`
	Image            string            `json:"image" mapstructure:"image"`
	Environment      map[string]string `json:"environment" mapstructure:"environment"`
	WorkingDirectory string            `json:"workingDirectory" mapstructure:"workingDirectory"`
	TimeoutSeconds   int               `json:"timeoutSeconds" mapstructure:"timeoutSeconds"`
	DisableCleanup   bool              `json:"disableCleanup" mapstructure:"disableCleanup"`
}

// Produce should produce Artifact
func (t *Target) Produce(checksum hash.Hash) (ark.Artifact, error) {
	return &Artifact{
		RawArtifact: ark.RawArtifact{
			Key:        t.Key(),
			Type:       t.Type,
			Hash:       hex.EncodeToString(checksum.Sum(nil)),
			Attributes: nil,
		},
	}, nil
}

// Validate checks if the Target fields are valid
func (t *Target) Validate() error {
	if err := t.RawTarget.Validate(); err != nil {
		return err
	}
	return validation.ValidateStruct(t,
		validation.Field(&t.Args, validation.Required),
		validation.Field(&t.Image, validation.Required))
}
