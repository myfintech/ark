package ark

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/myfintech/ark/src/go/lib/utils/cloudutils"

	"github.com/stretchr/testify/require"
)

func TestParseKey(t *testing.T) {
	t.Run("correct key format", func(t *testing.T) {
		key := "foo:bar"
		result, err := ParseKey(key)
		require.NoError(t, err)
		require.Equal(t, Key{
			Path: "foo",
			Name: "bar",
		}, result)
	})
	t.Run("key too short", func(t *testing.T) {
		key := "baz"
		_, err := ParseKey(key)
		require.Error(t, err)
	})
	t.Run("key too long", func(t *testing.T) {
		key := "foo:bar:baz"
		_, err := ParseKey(key)
		require.Error(t, err)
	})
}

func TestRawArtifact(t *testing.T) {
	mockRawArtifact := RawArtifact{
		Key:                "foo/build.ts:bar",
		Hash:               "1234567890",
		RemoteCacheBaseURL: "gs://ark-cache",
	}

	cacheDir, err := mockRawArtifact.MkCacheDir()
	require.NoError(t, err)
	require.DirExists(t, cacheDir)

	testFileName := filepath.Join(cacheDir, "foo.txt")
	require.NoError(t, os.WriteFile(testFileName, []byte("This is a file"), 0644))
	require.NoError(t, mockRawArtifact.WriteState())
	require.FileExists(t, testFileName)

	require.NoError(t, mockRawArtifact.Push(context.TODO()))
	exists, err := mockRawArtifact.RemotelyCached(context.TODO())
	require.NoError(t, err)
	require.True(t, exists)

	require.NoError(t, os.RemoveAll(cacheDir))
	exists, err = mockRawArtifact.LocallyCached(context.TODO())
	require.Error(t, err)
	require.False(t, exists)

	require.NoError(t, mockRawArtifact.Pull(context.TODO()))
	require.FileExists(t, testFileName)

	exists, err = mockRawArtifact.LocallyCached(context.TODO())
	require.NoError(t, err)
	require.True(t, exists)

	require.NoError(t, cloudutils.DeleteBlob(context.TODO(), mockRawArtifact.RemoteCacheBaseURL, fmt.Sprintf("%s.tar.gz", mockRawArtifact.Hash)))
}
