package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gitignorev4 "gopkg.in/src-d/go-git.v4/plumbing/format/gitignore"

	"github.com/stretchr/testify/require"
)

func TestReadFileBytes(t *testing.T) {
	testdataDir, err := testingSetup()
	require.NoError(t, err)

	_, err = ReadFileBytes(filepath.Join(testdataDir, "keep_file_1.md"))
	require.NoError(t, err)
}

func TestReadFileString(t *testing.T) {
	testdataDir, err := testingSetup()
	require.NoError(t, err)

	_, err = ReadFileString(filepath.Join(testdataDir, "keep_file_1.md"))
	require.NoError(t, err)
}

func TestFileDirectory(t *testing.T) {
	testdataDir, err := testingSetup()
	require.NoError(t, err)

	cwd, err := os.Getwd()
	require.NoError(t, err)

	require.Equal(t, filepath.Join(cwd, "testdata"), filepath.Dir(filepath.Join(testdataDir, "keep_file_1.md")))
}

func TestCompareFileHash(t *testing.T) {
	testdataDir, err := testingSetup()
	require.NoError(t, err)

	expectedHash := "a305b5e03176b3567cc86af01d167a4c2868f5aa81eb24b2a2db7a6bcdb7797a"

	matched, err := CompareFileHash(filepath.Join(testdataDir, "keep_dir_1", "keep_file_2.md"), expectedHash)
	require.NoError(t, err)
	require.True(t, matched)
}

func testingSetup() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, "testdata"), nil
}

func TestGlob(t *testing.T) {
	testdataDir, err := testingSetup()
	require.NoError(t, err)

	patterns := []gitignorev4.Pattern{
		gitignorev4.ParsePattern("do_not", []string{}),
		gitignorev4.ParsePattern("keep_file_2.md", []string{}),
	}

	shouldBeIgnored := []string{
		".gitkeep",
		"keep_dir_1",
		"keep_dir_2",
		"keep_file_2.md",
		"do_not/traverse/file.txt",
	}

	shouldContain := []string{
		filepath.Join(testdataDir, "keep_file_1.md"),
	}

	globs := []struct {
		name    string
		pattern string
	}{
		{"single star", "*"},
		{"double star", "**/*"},
	}
	for _, glob := range globs {
		t.Run(fmt.Sprintf("should match %s", glob.name), func(t *testing.T) {
			matches, err := Glob(
				glob.pattern,
				testdataDir,
				gitignorev4.NewMatcher(patterns),
				FilterDirectories,
				FilterBySuffix(".gitkeep"),
			)

			require.NoError(t, err)
			require.NotEmpty(t, matches)
			require.Equal(t, shouldContain, matches)
			for _, match := range matches {
				for _, suffix := range shouldBeIgnored {
					if strings.HasSuffix(match, suffix) {
						t.Errorf("%s should have been ignored", match)
					}
				}
			}
		})
	}
}

func TestGlobErrors(t *testing.T) {
	testdataDir, err := testingSetup()
	require.NoError(t, err)
	_, err = Glob("abc/**/*", testdataDir, nil)
	require.NoError(t, err)
}
