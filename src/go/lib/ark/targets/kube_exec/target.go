package kube_exec

import (
	"encoding/hex"
	"hash"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/myfintech/ark/src/go/lib/ark"
)

// Type is the string value of the Target type
const Type = "kube_exec"

// Target expresses the intention to implement a kube_exec target
type Target struct {
	ark.RawTarget  `mapstructure:",squash"`
	ResourceType   string   `json:"resourceType" mapstructure:"resourceType"`
	ResourceName   string   `json:"resourceName" mapstructure:"resourceName"`
	Command        []string `json:"command" mapstructure:"command"`
	ContainerName  string   `json:"containerName" mapstructure:"containerName"`
	TimeoutSeconds int      `json:"timeoutSeconds" mapstructure:"timeoutSeconds"`
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

// // Type returns the target type
// func (t Target) Type() string {
//	return Type
// }

// Produce should produce a deterministic kube_exec artifact
// func (t Target) Produce() (ark.Artifact, error) {
//	hash, err := ark.Hash(t)
//	if err != nil {
//		return nil, err
//	}
//
//	hexHash := hex.EncodeToString(hash.Sum(nil))
//
//	return &Artifact{
//		Key:  t.Key(),
//		Hash: hexHash,
//	}, nil
// }

// Validate checks if the Target fields are valid
func (t *Target) Validate() error {
	if err := t.RawTarget.Validate(); err != nil {
		return err
	}
	return validation.ValidateStruct(t,
		validation.Field(&t.ResourceType, validation.Required),
		validation.Field(&t.ResourceName, validation.Required),
		validation.Field(&t.Command, validation.Required),
	)
}
