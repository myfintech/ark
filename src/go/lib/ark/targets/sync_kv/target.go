package sync_kv

import (
	"encoding/hex"
	"hash"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/myfintech/ark/src/go/lib/ark"
)

// Type is the string value of the Target type
const Type = "sync_kv"

// Target expresses the intention to implement a KV Sync target
type Target struct {
	ark.RawTarget  `mapstructure:",squash"`
	Engine         string `json:"engine" mapstructure:"engine"`
	EngineURL      string `json:"engineUrl" mapstructure:"engineUrl"`
	TimeoutSeconds int    `json:"timeoutSeconds" mapstructure:"timeoutSeconds"`
	Token          string `json:"token" mapstructure:"token"`
	MaxRetries     int    `json:"maxRetries" mapstructure:"maxRetries"`
}

// TODO: review if we can extract a defaultFN for an artifact.

// Produce should produce Artifact
func (t Target) Produce(checksum hash.Hash) (ark.Artifact, error) {
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
		validation.Field(&t.Engine, validation.Required),
		validation.Field(&t.EngineURL, validation.Required),
		validation.Field(&t.Token, validation.Required),
		validation.Field(&t.SourceFiles, validation.Required),
	)
}
