package local_file

import (
	"encoding/hex"
	"hash"
	"path/filepath"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/myfintech/ark/src/go/lib/ark"
)

// Type is the string value of the Target type
const Type = "local_file"

// Target expresses an intention to write a file to disk
type Target struct {
	ark.RawTarget `mapstructure:",squash"`
	Filename      string `json:"filename" mapstructure:"filename"`
	Content       string `json:"content" mapstructure:"content"`
}

// Produce should produce Artifact
func (t *Target) Produce(checksum hash.Hash) (ark.Artifact, error) {
	shasum := hex.EncodeToString(checksum.Sum(nil))

	artifact := &Artifact{
		RenderedFilePath: t.Filename,
		RawArtifact: ark.RawArtifact{
			Key:        t.Key(),
			Type:       t.Type,
			Hash:       shasum,
			Attributes: nil,
		},
	}

	if !filepath.IsAbs(artifact.RenderedFilePath) {
		cacheDir, err := artifact.CacheDirPath()
		if err != nil {
			return artifact, err
		}
		artifact.RenderedFilePath = filepath.Join(cacheDir, artifact.RenderedFilePath)
	}

	return artifact, nil
}

// Validate checks if the Target fields are valid
func (t *Target) Validate() error {
	if err := t.RawTarget.Validate(); err != nil {
		return err
	}
	return validation.ValidateStruct(t,
		validation.Field(&t.Filename, validation.Required),
		validation.Field(&t.Content, validation.Required),
	)
}
