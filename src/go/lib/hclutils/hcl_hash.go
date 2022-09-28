package hclutils

import (
	"crypto/sha256"
	"hash"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// BlockMatcher an interface that accepts an HCL file and queries it for matching blocks
type BlockMatcher func(file *hclwrite.File) *hclwrite.Block

// IsolateBlocks takes source HCL bytes and a blockMatcher to isolate matching blocks to a single file (useful for hashing)
func IsolateBlocks(sourceBytes []byte, matcher BlockMatcher, fileName string) (*hclwrite.File, error) {
	sourceHCLFile, diag := hclwrite.ParseConfig(sourceBytes, fileName, hcl.InitialPos)
	if diag != nil && diag.HasErrors() {
		return sourceHCLFile, diag
	}

	blocks := matcher(sourceHCLFile)
	isolatedBlockFile := hclwrite.NewEmptyFile()
	isolatedBlockFile.Body().AppendBlock(blocks)
	return isolatedBlockFile, nil
}

// HashFile a convenience method for writing bytes to a hasher.
// if a hasher isn't provided it defaults to sha256
func HashFile(file *hclwrite.File, hasher hash.Hash) (hash.Hash, error) {
	if hasher == nil {
		hasher = sha256.New()
	}
	_, err := file.WriteTo(hasher)
	return hasher, err
}

// Isolator an interface describing an object that can isolate itself
type Isolator interface {
	IsolateHCLBlocks() (*hclwrite.File, error)
}
