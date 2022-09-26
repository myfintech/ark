package probe

import (
	"encoding/hex"
	"hash"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/myfintech/ark/src/go/lib/ark"
)

// Type is the string value of the Target type
const Type = "probe"

// Target expresses the intention to implement a Probe target
type Target struct {
	ark.RawTarget  `mapstructure:",squash"`
	DialAddress    string `json:"address" mapstructure:"address"`
	Timeout        string `json:"timeout" mapstructure:"timeout"`
	Delay          string `json:"delay" mapstructure:"delay"`
	MaxRetries     int    `json:"maxRetries" mapstructure:"maxRetries"`
	ExpectedStatus int    `json:"expectedStatus" mapstructure:"expectedStatus"`
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
		validation.Field(&t.DialAddress, validation.Required),
	)
}
