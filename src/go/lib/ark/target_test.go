package ark

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBaseTarget(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	testdata := filepath.Join(cwd, "testdata")

	target := RawTarget{
		Name:  "example",
		Type:  "test",
		Realm: cwd,
		File:  filepath.Join(cwd, "target_test.go"),
		SourceFiles: []string{
			filepath.Join(testdata, "01_dont_change_me.txt"),
			filepath.Join(testdata, "02_dont_change_me.txt"),
		},
	}

	err = target.Validate()
	require.NoError(t, err)

	hash, err := target.Checksum()
	require.NoError(t, err)
	require.Equal(t, "8c8932a81125a10ca86e2408f0c5d10c9b232843062438bf9c68bdb0f5de28fa", hex.EncodeToString(hash.Sum(nil)))
}
