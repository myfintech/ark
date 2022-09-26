package docker_image

import (
	"encoding/hex"
	"fmt"
	"hash"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/myfintech/ark/src/go/lib/ark"
)

// Type is the string value of the Target type
const Type = "docker_image"

// Target expresses intention to build a docker image
type Target struct {
	ark.RawTarget              `mapstructure:",squash"`
	Repo                       string             `json:"repo" mapstructure:"repo"`
	Dockerfile                 string             `json:"dockerfile" mapstructure:"dockerfile"`
	Secrets                    []string           `json:"secrets" mapstructure:"secrets"`
	DisableEntrypointInjection bool               `json:"disableEntrypointInjection" mapstructure:"disableEntrypointInjection"`
	CacheInline                bool               `json:"cacheInline" mapstructure:"cacheInline"`
	BuildArgs                  map[string]*string `json:"BuildArgs" mapstructure:"buildArgs"`
	Output                     string             `json:"output" mapstructure:"output"`
	CacheFrom                  []string           `json:"cacheFrom" mapstructure:"cacheFrom"`
}

// Produce should produce a deterministic docker image artifact
func (t Target) Produce(checksum hash.Hash) (ark.Artifact, error) {
	artifact := &Artifact{
		RawArtifact: ark.RawArtifact{
			Key:  t.Key(),
			Type: t.Type,
			Hash: hex.EncodeToString(checksum.Sum(nil)),
			Attributes: map[string]interface{}{
				"secrets": t.Secrets,
			},
		},
	}
	artifact.URL = fmt.Sprintf("%s:%s", t.Repo, artifact.Hash)
	return artifact, nil
}

// Validate checks if the Target fields are valid
func (t *Target) Validate() error {
	if err := t.RawTarget.Validate(); err != nil {
		return err
	}

	return validation.ValidateStruct(t,
		validation.Field(&t.Repo, validation.Required),
		validation.Field(&t.Dockerfile, validation.Required),
	)
}
