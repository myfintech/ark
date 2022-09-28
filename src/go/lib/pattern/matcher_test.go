package pattern

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatcher(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	fileMatcher := &Matcher{}
	testdataDir := filepath.Join(cwd, "testdata")
	require.NoError(t, fileMatcher.Compile())

	matchedFiles, _, err := fileMatcher.Walk([]string{testdataDir})
	require.NoError(t, err)

	for _, file := range matchedFiles {
		t.Log(file)
	}

	t.Run("should be able to include a single file", func(t *testing.T) {
		fileMatcher = &Matcher{
			Paths:    []string{"/example/dir"},
			Includes: []string{"/example/dir/file-1.txt"},
		}
		require.NoError(t, fileMatcher.Compile())
		require.False(t, fileMatcher.Check("/not/my/File"), "/not/my/File should not be included")
		require.False(t, fileMatcher.Check("/not/my/file-1.txt"), "/not/my/file-1.txt should not be included")
		require.True(t, fileMatcher.Check("/example/dir/file-1.txt"), "/example/dir/file-1.txt should be included")
		require.False(t, fileMatcher.Check("/example/dir/file-2.txt"), "/example/dir/file-2.txt should not be included")
	})

	t.Run("should be able to include all files in a directory", func(t *testing.T) {
		fileMatcher = &Matcher{
			Paths:    []string{"/example/dir"},
			Includes: []string{"**/other_dir/**"},
		}
		require.NoError(t, fileMatcher.Compile())
		require.False(t, fileMatcher.Check("/not/my/File"), "/not/my/File should not be included")
		require.False(t, fileMatcher.Check("/example/dir/file-1.txt"), "/example/dir/file-1.txt should not be included")
		require.True(t, fileMatcher.Check("/example/dir/other_dir/file-1.txt"), "/example/dir/other_dir/file-1.txt should be included")
		require.True(t, fileMatcher.Check("/example/dir/other_dir/file-2.txt"), "/example/dir/other_dir/file-2.txt should be included")
		require.True(t, fileMatcher.Check("/example/dir/other_dir/file-a.txt"), "/example/dir/other_dir/file-a.txt should be included")
		require.True(t, fileMatcher.Check("/example/dir/other_dir/file-A.txt"), "/example/dir/other_dir/file-A.txt should be included")
	})

	t.Run("should be able to exclude files from a directory", func(t *testing.T) {
		fileMatcher = &Matcher{
			Paths:    []string{"/example/dir"},
			Excludes: []string{"**/file-[a-z].txt"},
		}
		require.NoError(t, fileMatcher.Compile())
		require.False(t, fileMatcher.Check("/not/my/File"), "/not/my/File should not be included")
		require.True(t, fileMatcher.Check("/example/dir/other_dir/file-1.txt"), "/example/dir/other_dir/file-1.txt should be included")
		require.True(t, fileMatcher.Check("/example/dir/other_dir/file-2.txt"), "/example/dir/other_dir/file-2.txt should be included")
		require.False(t, fileMatcher.Check("/example/dir/other_dir/file-a.txt"), "/example/dir/other_dir/file-a.txt should not be included")
		require.True(t, fileMatcher.Check("/example/dir/other_dir/file-A.txt"), "/example/dir/other_dir/file-A.txt should be included")
	})

	t.Run("should be able to include only certain files from a directory", func(t *testing.T) {
		fileMatcher = &Matcher{
			Paths:    []string{"/example/dir"},
			Includes: []string{"/example/dir/other_dir/file-[0-9].txt"},
			Excludes: []string{"**/file-[a-z].txt"},
		}
		require.NoError(t, fileMatcher.Compile())
		require.False(t, fileMatcher.Check("/not/my/File"), "/not/my/File should not be included")
		require.True(t, fileMatcher.Check("/example/dir/other_dir/file-1.txt"), "/example/dir/other_dir/file-1.txt should be included")
		require.True(t, fileMatcher.Check("/example/dir/other_dir/file-2.txt"), "/example/dir/other_dir/file-2.txt should be included")
		require.False(t, fileMatcher.Check("/example/dir/other_dir/file-a.txt"), "/example/dir/other_dir/file-a.txt should not be included")
		require.False(t, fileMatcher.Check("/example/dir/other_dir/file-A.txt"), "/example/dir/other_dir/file-A.txt should not be included")
	})

	t.Run("should be able to include only certain patterns without a base path", func(t *testing.T) {
		fileMatcher = &Matcher{
			Includes: []string{"**/other_dir/file-[0-9].txt"},
		}
		require.NoError(t, fileMatcher.Compile())
		require.False(t, fileMatcher.Check("/not/my/File"), "/not/my/File should not be included")
		require.True(t, fileMatcher.Check("/example/dir/other_dir/file-1.txt"), "/example/dir/other_dir/file-1.txt should be included")
		require.True(t, fileMatcher.Check("/example/dir/other_dir/file-2.txt"), "/example/dir/other_dir/file-2.txt should be included")
		require.False(t, fileMatcher.Check("/example/dir/other_dir/file-a.txt"), "/example/dir/other_dir/file-a.txt should not be included")
		require.False(t, fileMatcher.Check("/example/dir/other_dir/file-A.txt"), "/example/dir/other_dir/file-A.txt should not be included")
	})

	t.Run("an empty matcher should never return a match", func(t *testing.T) {
		fileMatcher = &Matcher{}
		require.NoError(t, fileMatcher.Compile())
		require.False(t, fileMatcher.Check("/not/my/File"), "/not/my/File should be included")
		require.False(t, fileMatcher.Check("/example/dir/other_dir/file-1.txt"), "/example/dir/other_dir/file-1.txt should be included")
		require.False(t, fileMatcher.Check("/example/dir/other_dir/file-2.txt"), "/example/dir/other_dir/file-2.txt should be included")
		require.False(t, fileMatcher.Check("/example/dir/other_dir/file-a.txt"), "/example/dir/other_dir/file-a.txt should be included")
		require.False(t, fileMatcher.Check("/example/dir/other_dir/file-A.txt"), "/example/dir/other_dir/file-A.txt should be included")
	})
}
