package jsonnetlib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstall(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err, "there should not be an error getting the current working directory")

	testData := filepath.Join(cwd, "testdata")
	checksumFile := filepath.Join(testData, "checksum")

	defer func() {
		_ = os.RemoveAll(testData)
	}()

	err = Install(testData)
	require.NoError(t, err, "there should not be an error running the Install function")
	require.DirExists(t, testData, "the directory should exist")
	require.DirExists(t, filepath.Join(testData, "k8s"), "the directory should exist")
	require.DirExists(t, filepath.Join(testData, "microservice"), "the directory should exist")
	require.FileExists(t, checksumFile)

	info, _ := os.Stat(checksumFile)
	modTime := info.ModTime()
	err = Install(testData)
	require.NoError(t, err, "there should not be an error running the Install function")

	info, _ = os.Stat(checksumFile)
	require.Equal(t, modTime, info.ModTime(), "modTime should not have changed")
}
