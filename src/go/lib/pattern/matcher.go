package pattern

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"
)

// Matcher uses patterns to filter files
type Matcher struct {
	Paths            []string
	Includes         []string
	Excludes         []string
	CompiledIncludes []glob.Glob
	CompiledExcludes []glob.Glob
}

// Compile creates globs for each include/exclude pattern provided
func (m *Matcher) Compile() error {
	m.CompiledIncludes = nil
	m.CompiledExcludes = nil

	for _, pattern := range m.Includes {
		g, err := glob.Compile(strings.TrimSpace(pattern))
		if err != nil {
			return errors.Wrapf(err, "failed to compile pattern --> %s", pattern)
		}
		m.CompiledIncludes = append(m.CompiledIncludes, g)
	}

	for _, pattern := range m.Excludes {
		g, err := glob.Compile(strings.TrimSpace(pattern))
		if err != nil {
			return errors.Wrapf(err, "failed to compile pattern --> %s", pattern)
		}
		m.CompiledExcludes = append(m.CompiledExcludes, g)
	}

	return nil
}

// MustCompile is like Compile but panics if the provided expressions cannot be parsed.
// It simplifies safe initialization of global variables holding compiled regular expressions.
func (m *Matcher) MustCompile() {
	if err := m.Compile(); err != nil {
		panic(err)
	}
}

// HasPrefix checks if a file matches any of the provides paths
func (m *Matcher) HasPrefix(path string) bool {
	for _, pattern := range m.Paths {
		if strings.HasPrefix(path, pattern) {
			return true
		}
	}
	return false
}

// Included checks if a file matches an inclusion pattern
func (m *Matcher) Included(path string) bool {
	for _, pattern := range m.CompiledIncludes {
		if pattern.Match(path) {
			return true
		}
	}
	return false
}

// Excluded checks if a file matches any of the exclusion patterns
func (m *Matcher) Excluded(path string) bool {
	for _, pattern := range m.CompiledExcludes {
		if pattern.Match(path) {
			return true
		}
	}
	return false
}

// Check returns true if the string is included and not excluded from the matcher
func (m *Matcher) Check(path string) bool {

	// This is an no data was provided to match
	if len(m.Paths) == 0 && len(m.Includes) == 0 && len(m.Excludes) == 0 {
		return false
	}

	// comparing paths with a prefix against includes
	if len(m.Paths) > 0 && len(m.CompiledIncludes) > 0 {
		return m.HasPrefix(path) && m.Included(path) && !m.Excluded(path)
	}

	// comparing paths with a prefix against only excludes because no includes present
	if len(m.Paths) > 0 && len(m.CompiledIncludes) == 0 {
		return m.HasPrefix(path) && !m.Excluded(path)
	}

	// comparing against inclusions and exclusions
	if len(m.Paths) == 0 && len(m.CompiledIncludes) > 0 {
		return m.Included(path) && !m.Excluded(path)
	}
	return !m.Excluded(path)
}

// Some uses Check to validate that at least one of the strings in the provided paths match
func (m *Matcher) Some(paths []string) bool {
	for _, path := range paths {
		if m.Check(path) {
			return true
		}
	}
	return false
}

/*
Deprecated: Walk walks the specified files to determine matches against it's declared patterns.
If files provided contain directories it will traverse all files and sub directories.
The files viewed in lexical order, which makes the output deterministic but means that for very
large directories Walk can be inefficient.
Walk does not follow symbolic links.
*/
func (m *Matcher) Walk(files []string) ([]string, map[string]bool, error) {
	var matchedFiles []string
	matchedFilesMap := map[string]bool{}

	sort.Strings(files)

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			return matchedFiles, matchedFilesMap, errors.Wrapf(err, "failed to state: %s", file)
		}

		if info.IsDir() {
			err = filepath.Walk(file, func(path string, info os.FileInfo, _ error) error {
				if info.IsDir() {
					return nil
				}

				if m.Included(path) && !m.Excluded(path) {
					matchedFiles = append(matchedFiles, path)
					matchedFilesMap[path] = true
				}

				return nil
			})

			if err != nil {
				return matchedFiles, matchedFilesMap, errors.Wrapf(err, "failed to walk %s", file)
			}

			continue
		}

		if m.Included(file) && !m.Excluded(file) {
			matchedFiles = append(matchedFiles, file)
			matchedFilesMap[file] = true
		}
	}

	return matchedFiles, matchedFilesMap, nil
}
