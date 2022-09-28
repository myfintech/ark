package cloudutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCloudUtils(t *testing.T) {
	tempFile := "source.txt"
	tempSourceFileDir := "/tmp/cloudutils_testing/source"
	tempSourceFilePath := filepath.Join(tempSourceFileDir, tempFile)
	tempDestFileDir := "/tmp/cloudutils_testing/destination"
	tempDestFilePath := filepath.Join(tempDestFileDir, tempFile)
	blobURL := strings.Join([]string{"file://", tempDestFileDir}, "")

	require.NoError(t, os.MkdirAll(tempSourceFileDir, 0755), "should be able to create source directories")

	sourceWriter, err := os.OpenFile(tempSourceFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	require.NoError(t, err, "should be able to create a test source file")

	_, err = fmt.Fprintln(sourceWriter, "This is a test")
	require.NoError(t, err, "should be able to write to source temp file")
	require.NoError(t, sourceWriter.Close(), "should be able to close source temp file")

	err = os.MkdirAll(tempDestFileDir, 0755)
	require.NoError(t, os.MkdirAll(tempDestFileDir, 0755), "should be able to create destination directories")

	destWriter, err := os.OpenFile(tempDestFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	require.NoError(t, err, "should be able to create a test destination file")

	_, err = fmt.Fprintln(destWriter, "This is also a test")
	require.NoError(t, err, "should be able to write to destination temp file")
	require.NoError(t, destWriter.Close(), "should be able to close destination temp file")

	t.Run("should be able to create a writer for a remote location", func(t *testing.T) {
		_, cleanup, blobWriteErr := NewBlobWriter(nil, blobURL, tempFile)
		require.NoError(t, blobWriteErr, "should be able to get a writer")
		cleanup()
	})
	t.Run("should be able to verify that file exists in a location", func(t *testing.T) {
		exists, blobCheckErr := BlobCheck(nil, blobURL, tempFile)
		require.NoError(t, blobCheckErr, "should be able to verify existence of file")
		require.True(t, exists, "destination temp file should exist")
	})
	t.Run("should be able to create a reader for a remote location", func(t *testing.T) {
		_, cleanup, blobReadErr := NewBlobReader(nil, blobURL, tempFile)
		require.NoError(t, blobReadErr, "should be able to get a reader")
		cleanup()
	})

	t.Run("should be able to delete a blob", func(t *testing.T) {
		require.NoError(t, DeleteBlob(nil, blobURL, tempFile))
	})

	require.NoError(t, os.RemoveAll("/tmp/cloudutils_testing"), "should be able to cleanup after tests")
}
