package ark

import (
	"context"
	"hash"
)

// Hashable is an interface implemented by structs that can produce a deterministic hash of their properties
type Hashable interface {
	Hash() (hash.Hash, error)
}

// Validatable is an interface implemented by a struct who's fields may be validated
type Validatable interface {
	Validate() error
}

// Producer is an interface that is implemented by a struct that produces an Artifact
type Producer interface {
	Produce(checksum hash.Hash) (Artifact, error)
}

// Action is an interface is implemented by a struct that should use a RawTarget to configure its execution and produce an Artifact
type Action interface {
	Execute(ctx context.Context) error
}

type Target interface {
	Producer
	Validatable
	Key() string
}

type Derivation struct {
	Target   Target
	Artifact Artifact
}

type Derivative struct {
	RawTarget   RawTarget   `json:"Target"`
	RawArtifact RawArtifact `json:"Artifact"`
}
