package nix

import (
	"encoding/hex"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/myfintech/ark/src/go/lib/ark"
	"hash"
	"strings"
)

// Type is the string value of the Target type
const Type = "nix"

// Target expresses the intention to install one or more nix packages
type Target struct {
	ark.RawTarget `mapstructure:",squash"`
	Packages      []string `json:"packages" mapstructure:"packages"`
}

// Produce should produce Artifact
func (t *Target) Produce(checksum hash.Hash) (ark.Artifact, error) {
	artifact := &Artifact{
		RawArtifact: ark.RawArtifact{
			Key:        t.Key(),
			Type:       t.Type,
			Hash:       hex.EncodeToString(checksum.Sum(nil)),
			Attributes: nil,
		},
	}

	// go ahead and get the proper name for the packages here rather than doing the string splitting elsewhere
	// it increases readability in whatever datastore and makes future actions simpler (local cache checks)
	pkgNames := make([]string, 0)
	var pkgName string
	for _, pkg := range t.Packages {
		splitStr := strings.Split(pkg, ".")
		pkgName = splitStr[len(splitStr)-1]
		pkgNames = append(pkgNames, pkgName)
	}

	artifact.Packages = pkgNames
	return artifact, nil
}

// Validate checks if the Target fields are valid
func (t *Target) Validate() error {
	if err := t.RawTarget.Validate(); err != nil {
		return err
	}
	return validation.ValidateStruct(t,
		validation.Field(&t.Packages, validation.Required),
	)
}
