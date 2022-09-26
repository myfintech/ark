package hclutils

import (
	"encoding/hex"
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/stretchr/testify/require"
)

const hashableHCL = `
package "skip_me" {
	description = "example"
}

target "build" "other" {}

target "build" "example" {
	repo = "gcr.io"
}
`

func TestHash(t *testing.T) {
	isolatedBlocks, err := IsolateBlocks([]byte(hashableHCL), func(file *hclwrite.File) *hclwrite.Block {
		return file.Body().FirstMatchingBlock("target", []string{"build", "example"})
	}, "example")
	require.NoError(t, err)

	hash, err := HashFile(isolatedBlocks, nil)
	require.NoError(t, err)
	require.Equal(t, "3632d0d8273018cfd0dcb034381bc0c7a9120e5c559428d22524dd30e67d9911", hex.EncodeToString(hash.Sum(nil)))
}
