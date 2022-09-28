package derivation

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/ark/targets/docker_image"
)

func TestDeriveRawArtifactFromRawTarget(t *testing.T) {
	rawTarget := ark.RawTarget{
		Name:  "test",
		Type:  docker_image.Type,
		File:  "example",
		Realm: "",
		Attributes: map[string]interface{}{
			"repo":       "example",
			"dockerfile": "FROM node:latest",
		},
		SourceFiles: nil,
		DependsOn:   nil,
	}

	rawArtifact, err := RawArtifactFromRawTarget(rawTarget)
	require.NoError(t, err)
	t.Log(rawArtifact)
}
