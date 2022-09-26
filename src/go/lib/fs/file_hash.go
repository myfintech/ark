package fs

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// HashFile calculates the sha256 of the given files contents (defaults to sha256 hash)
func HashFile(file string, rootHash hash.Hash) (hash.Hash, error) {
	if rootHash == nil {
		rootHash = sha256.New()
	}

	fileHandle, err := os.Open(file)
	if err != nil {
		return rootHash, err
	}
	defer fileHandle.Close()

	// stream the contents of the file through the file hasher
	_, err = io.Copy(rootHash, fileHandle)

	if err != nil {
		return rootHash, err
	}

	// Write the file hash and file name to the dir hash io.writer
	return rootHash, nil
}

// HashDir calculates the root hash of all files in a given directory (defaults to sha256 hash)
func HashDir(dir string, rootHash hash.Hash) (hash.Hash, error) {
	if rootHash == nil {
		rootHash = sha256.New()
	}

	files, err := SortedFiles(filepath.Clean(dir))

	if err != nil {
		return rootHash, err
	}

	for _, file := range files {
		// we use the filename as a part of the rootHash which is new line delimited
		if strings.Contains(file, "\n") {
			return rootHash, errors.New("dirhash: filenames cannot contain new lines")
		}

		fileHash, hashErr := HashFile(file, nil)

		if hashErr != nil {
			return rootHash, hashErr
		}

		// Write the file hash and file name to the dir hash io.writer
		fmt.Fprintf(rootHash, "%x  %s\n", fileHash.Sum(nil), file)
	}

	return rootHash, nil
}

// HashFiles hashes all provides files for a root hash
