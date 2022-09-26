package gitignore

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	gitignorev5 "github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/fs"
	"github.com/myfintech/ark/src/go/lib/log"
)

const (
	commentPrefix = "#"
	eol           = "\n"
	ignoreFile    = ".gitignore"
	gitDir        = ".git"
)

type Pattern gitignorev5.Pattern

var ctxLog = log.WithFields(log.Fields{
	"prefix": "gitignore",
})

// NewMatcher constructs a new global matcher. Patterns must be given in the order of
// increasing priority. That is most generic settings files first, then the content of
// the repo .gitignore, then content of .gitignore down the path or the repo and then
// the content command line arguments.
func NewMatcher(p []gitignorev5.Pattern) gitignorev5.Matcher {
	return gitignorev5.NewMatcher(p)
}

// Domain parses the domain of a gitignore file by the base and the current path
func Domain(base, path string) ([]string, error) {
	if !strings.HasPrefix(path, base) {
		return nil, errors.Errorf("%s is not a prefix of %s", base, path)
	}

	domainBase := fs.Split(base)
	domainPath := fs.Split(path)

	return domainPath[len(domainBase):], nil
}

// LoadRepoPatterns reads gitignore patterns recursively traversing through the directory
// structure. The result is in the ascending order of priority (last higher).
func LoadRepoPatterns(repoDir string) (patterns []gitignorev5.Pattern, err error) {
	rootIgnore := filepath.Join(repoDir, ignoreFile)
	patterns, err = ReadIgnoreFile(rootIgnore, nil)
	matcher := gitignorev5.NewMatcher(patterns)
	if err != nil {
		return
	}
	err = filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		dirName := filepath.Dir(path)

		// ignore root ignore file
		if path == rootIgnore {
			ctxLog.Tracef("skipping %s", path)
			return nil
		}

		// don't crawl the .git directory
		if info.Name() == gitDir {
			ctxLog.Tracef("skipping %s", path)
			return filepath.SkipDir
		}

		// skip ignored directories for performance
		if info.IsDir() && matcher.Match(fs.Split(path), info.IsDir()) {
			ctxLog.Tracef("skipping ignored by root %s", path)
			return filepath.SkipDir
		}

		if info.Name() == ignoreFile {
			domain, domainErr := Domain(repoDir, dirName)
			if domainErr != nil {
				return domainErr
			}
			subPatterns, readErr := ReadIgnoreFile(path, domain)
			if readErr != nil {
				return readErr
			}
			if len(subPatterns) > 0 {
				patterns = append(patterns, subPatterns...)
			}
		}
		return nil
	})
	return
}

// ReadIgnoreFile reads a specific git ignore file.
func ReadIgnoreFile(ignoreFilePath string, domain []string) (ps []gitignorev5.Pattern, err error) {
	data, err := ioutil.ReadFile(ignoreFilePath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	for _, s := range strings.Split(string(data), eol) {
		if !strings.HasPrefix(s, commentPrefix) && len(strings.TrimSpace(s)) > 0 {
			ps = append(ps, gitignorev5.ParsePattern(s, domain))
		}
	}
	return
}

// ParsePattern parses a gitignore pattern string into the Pattern structure.
