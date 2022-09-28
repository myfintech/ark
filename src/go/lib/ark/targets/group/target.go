package group

import (
	"encoding/hex"
	"hash"

	"github.com/myfintech/ark/src/go/lib/ark"
)

// Type is the string value of the Target type
const Type = "group"

// Target expresses an intention to target a Group of Targets
type Target struct {
	ark.RawTarget `mapstructure:",squash"`
}

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
	return t.RawTarget.Validate()
}
