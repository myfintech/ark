package observer

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/myfintech/ark/src/go/lib/watchman"

	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/stretchr/testify/require"

	"github.com/myfintech/ark/src/go/lib/fs"
	"github.com/myfintech/ark/src/go/lib/pattern"
)

var (
	expectedFileHash   = "6a3a1eaa0edbb69dc35cb03fe163edc4fca2b514"
	expectedTargetHash = "e3cae20e65403c29f30d153d1ea867bd58a5623146dc725787ca26931e85e509"
)

func TestObserver(t *testing.T) {
	cwd, _ := os.Getwd()
	testdata := filepath.Join(cwd, "testdata")
	testfile := filepath.Join(testdata, "hash_test.txt")
	modeSwitches := []bool{false, true}
	logger := logz.New(context.Background())

	wm, err := watchman.Connect(context.Background(), 10)
	require.NoError(t, err)

	for _, isNativeMode := range modeSwitches {
		mode := "native"
		if !isNativeMode {
			mode = "watchman"
		}
		t.Run(fmt.Sprintf("test observer mode %s", mode), func(t *testing.T) {
			observer := NewObserver(isNativeMode, true, testdata, []string{}, nil, logger, wm)
			err := observer.AddFileMatcher("test", &pattern.Matcher{
				Paths:    []string{testdata},
				Includes: []string{"*.txt"},
			})
			require.NoError(t, err)

			go func() {
				item, iErr := observer.FileSystemStream.First().Get()
				require.NoError(t, iErr)
				require.NoError(t, item.E)
			}()
			<-observer.WaitForInitialScan

			match, ok := observer.GetMatchCache("test")
			require.Equal(t, true, ok)
			require.NotNil(t, match)
			require.Equal(t, expectedTargetHash, hex.EncodeToString(match.Hash.Sum(nil)))

			record, ok := match.Files.Load(testfile)
			require.Equal(t, true, ok)

			require.Equal(t, expectedFileHash, record.(*fs.File).Hash)
		})
	}
	t.Run("test watch disabled", func(t *testing.T) {
		observer := NewObserver(true, false, testdata, []string{}, nil, logger, wm)
		err := observer.AddFileMatcher("test", &pattern.Matcher{
			Paths:    []string{testdata},
			Includes: []string{"*.txt"},
		})
		require.NoError(t, err)

		go func() {
			item, iErr := observer.FileSystemStream.First().Get()
			require.NoError(t, iErr)
			require.NoError(t, item.E)
		}()
		<-observer.WaitForInitialScan

		match, ok := observer.GetMatchCache("test")
		require.Equal(t, true, ok)
		require.NotNil(t, match)
		require.Equal(t, expectedTargetHash, hex.EncodeToString(match.Hash.Sum(nil)))

		record, ok := match.Files.Load(testfile)
		require.Equal(t, true, ok)

		require.Equal(t, expectedFileHash, record.(*fs.File).Hash)
	})
}
