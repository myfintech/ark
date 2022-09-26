package fs

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/log"
)

// ReadFileBytes an alias to ioutil.ReadFIle
func ReadFileBytes(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

// ReadFileString reads a files contents as a string
func ReadFileString(filename string) (string, error) {
	fileBytes, err := ReadFileBytes(filename)
	if err != nil {
		return "", err
	}
	return string(fileBytes), nil
}

// ReadFileJSON reads a file and unmarshalls the data into the given interface
func ReadFileJSON(filename string, v interface{}) error {
	data, err := ReadFileBytes(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// PathJoin creates a proper file path for both URLs and local filesystem paths

// SortedFiles aggregates all files under a specified directory and sorts lexically sorts them
func SortedFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(filepath.Clean(dir), func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		files = append(files, file)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// sort the files to ensure the hash is deterministic
	sort.Strings(files)
	return files, nil
}

// File a basic interface representing a file on disk
type File struct {
	Name          string
	Exists        bool
	New           bool
	Type          string
	Hash          string
	SymlinkTarget string
	RelName       string
}

// IsDir returns true of the file is a directory
func (f *File) IsDir() bool {
	return f.Type == "d"
}

// IsRegular returns true if the file is a regular file
func (f *File) IsRegular() bool {
	return f.Type == "f"
}

// TrimPrefix trims the given prefix off a filename
func TrimPrefix(file string, prefix string) string {
	return strings.TrimPrefix(strings.TrimPrefix(file, prefix), string(os.PathSeparator))
}

// TrimPrefixAll returns a sorted list of filenames with their prefix trimmed from the supplied list
func TrimPrefixAll(files []string, prefix string) []string {
	var trimmed []string
	for _, file := range files {
		trimmed = append(trimmed, TrimPrefix(file, prefix))
	}
	sort.Strings(trimmed)
	return trimmed
}

// Copy srcFile to destFile
func Copy(srcPath, destPath string) (err error) {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return
	}

	defer func() {
		_ = srcFile.Close()
	}()

	destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return
	}

	defer func() {
		_ = destFile.Close()
	}()

	if _, err = io.Copy(destFile, srcFile); err != nil {
		return
	}

	return
}

// CompareFileHash calculates a hash for a file and compares it against an expected hash
func CompareFileHash(sourcePath, expectedHash string) (bool, error) {
	hash, err := HashFile(sourcePath, nil)
	if err != nil {
		return false, err
	}

	actualHash := hex.EncodeToString(hash.Sum(nil))

	if expectedHash != actualHash {
		return false, errors.Errorf(
			"%s failed integrity check\nexpected %s\nrecieved %s",
			sourcePath,
			expectedHash,
			actualHash,
		)
	}
	return true, nil
}

// NativeFileToSynthetic accepts a path to a file and a os.FileInfo interface and provides a pointer to an fs.File
// This function can return errors when attempting to resolve a symlink location and when attempting to compute the file contents current hash
func NativeFileToSynthetic(root, path string, info os.FileInfo) (file *File, err error) {
	file = &File{
		Name:          path,
		Exists:        true,
		New:           false,
		Type:          "f",
		Hash:          "",
		SymlinkTarget: "",
		RelName:       TrimPrefix(path, root),
	}

	if info.IsDir() {
		file.Type = "d"
	}

	if info.Mode()&os.ModeSymlink == os.ModeSymlink {
		file.Type = "l"
		link, evalErr := filepath.EvalSymlinks(path)
		if evalErr != nil {
			return file, evalErr
		}
		file.SymlinkTarget = link
	}

	if info.Mode().IsRegular() {
		// using sha1 as the root hash to match facebook watchman
		rootHash, hErr := HashFile(path, sha1.New())
		if hErr != nil {
			return file, hErr
		}
		file.Hash = hex.EncodeToString(rootHash.Sum(nil))
	}
	return
}

// Split splits a given file path using the os.PathSeparator
func Split(path string) []string {
	p := strings.Split(path, string(os.PathSeparator))
	if len(p) > 0 && p[0] == "" {
		return p[1:]
	}
	return p
}

// NormalizePath expands a non absolute path by resolving home directory and ensure that the path is canonical for the host OS
func NormalizePath(baseDir, path string) (string, error) {
	// expand ${var} or $VAR
	path = os.ExpandEnv(path)

	// expand ~ to $HOME
	path, err := homedir.Expand(path)
	if err != nil {
		return "", fmt.Errorf("failed to expand home directory: %s", err)
	}

	// resolve relative to base
	if !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, path)
	}

	// clean any relative pathing .../../.
	return filepath.Clean(path), nil
}

// NormalizePathByPrefix normalizes files with relative or absolute prefixes
// / will be considered absolute
// ./ will be normalized to the relBase
// // will be normalized to the relBase
// otherwise they will be normalized to absBase
func NormalizePathByPrefix(file, absBase, relBase string) (cleanFile string, err error) {
	switch {
	case strings.HasPrefix(file, "./"):
		cleanFile, err = NormalizePath(relBase, file)
	case strings.HasPrefix(file, "//"):
		cleanFile, err = NormalizePath(absBase, strings.TrimPrefix(file, "//"))
	case filepath.IsAbs(file):
		cleanFile = file
	default:
		cleanFile, err = NormalizePath(absBase, file)
	}
	return
}

// FilterFiles filters the files list to matching files only

// FilterFunc should return true if a file should be retained
type FilterFunc func(file string, info os.FileInfo) bool

// Glob adds double-star support to the core path/filepath Glob function.
// It's useful when your globs might have double-stars, but you're not sure.
// You might recognize "**" recursive globs from things like your .gitignore file, and zsh.
// The "**" glob represents a recursive wildcard matching zero-or-more directory levels deep.
func Glob(pattern, realm string, matcher gitignore.Matcher, filters ...FilterFunc) (matches []string, err error) {
	pattern = strings.TrimSpace(pattern)
	pattern, err = NormalizePath(realm, pattern)
	if err != nil {
		return
	}

	// pass through to core package if no double-star
	if !strings.Contains(pattern, "**") {
		matches, err = filepath.Glob(pattern)
		if err != nil {
			return
		}
		return singleStarWalk(realm, matches, matcher, filters...)
	}

	chunks := strings.Split(pattern, "**")
	basePath, glob := chunks[0], chunks[1]

	err = filepath.Walk(basePath, doubleStarWalk(realm, matcher, filters, glob, func(path string) {
		matches = append(matches, path)
	}))

	return
}

func doubleStarWalk(realm string, matcher gitignore.Matcher, filters []FilterFunc, glob string, fn func(path string)) func(path string, info fs.FileInfo, err error) error {
	return func(path string, info fs.FileInfo, err error) error {
		if os.IsNotExist(err) {
			return nil
		}

		if err != nil {
			return err
		}

		ignored, err := excludeIgnoredFiles(path, realm, info, matcher, filters...)
		if err != nil {
			return err
		}

		if ignored {
			return nil
		}

		matched, err := filepath.Match(filepath.Dir(path)+glob, path)
		if err != nil {
			return err
		}

		if !matched {
			return nil
		}

		fn(path)

		return nil
	}
}

func singleStarWalk(realm string, files []string, matcher gitignore.Matcher, filters ...FilterFunc) ([]string, error) {
	var matches []string
	for _, file := range files {
		info, err := os.Stat(file)
		if os.IsNotExist(err) {
			return matches, nil
		}
		if err != nil {
			return matches, err
		}

		ignored, err := excludeIgnoredFiles(file, realm, info, matcher, filters...)
		if err == filepath.SkipDir {
			continue
		}
		if err != nil {
			return matches, err
		}
		if ignored {
			continue
		}
		matches = append(matches, file)
	}
	return matches, nil
}

func excludeIgnoredFiles(path string, realm string, info os.FileInfo, matcher gitignore.Matcher, filters ...FilterFunc) (ignored bool, err error) {
	if matcher != nil {
		ignored = matcher.Match(Split(TrimPrefix(path, realm)), info.IsDir())
	}

	if ignored && info.IsDir() {
		return ignored, filepath.SkipDir
	}

	if ignored {
		return
	}

	for _, filter := range filters {
		if !filter(path, info) {
			return true, nil
		}
	}
	return false, nil
}

// FilterBySuffix ignores files with the provided suffix
func FilterBySuffix(suffix string) FilterFunc {
	return func(file string, info os.FileInfo) bool {
		return !strings.HasSuffix(file, suffix)
	}
}

// FilterDirectories ignores directories in the file walk
func FilterDirectories(_ string, info os.FileInfo) bool {
	return !info.IsDir()
}
