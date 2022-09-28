package gitignore

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/myfintech/ark/src/go/lib/fs"
)

var (
	testBase = "/opt/app"
)

func TestDomain(t *testing.T) {
	t.Run("should return the correct suffix", func(t *testing.T) {
		domain, err := Domain(testBase, "/opt/app/example/domain")
		require.NoError(t, err)
		require.EqualValues(t, []string{"example", "domain"}, domain)
	})

	t.Run("should return an error when the base is not a prefix of the path", func(t *testing.T) {
		_, err := Domain(testBase, "/example/domain")
		require.Error(t, err)
	})
}

func TestLoadRepoPatterns(t *testing.T) {
	root, err := filepath.Abs("../../../../../")
	patterns, err := LoadRepoPatterns(root)
	require.NoError(t, err)
	require.NotEmpty(t, patterns)

	patternMatcher := NewMatcher(patterns)
	sqlFilePath := fs.Split("src/go/tools/monarch/seeds/sql/other/1531145774_random_1.sql")
	ignored := patternMatcher.Match(sqlFilePath, false)
	require.Equal(t, true, ignored, "monarch seed should be ignored")
}

func TestLoadRepoPatternsWithIgnore(t *testing.T) {
	root, err := filepath.Abs("../../../../../")
	patterns, err := LoadRepoPatterns(root)
	require.NoError(t, err)
	require.NotEmpty(t, patterns)

	patternMatcher := NewMatcher(patterns)
	buildDir := fs.Split("src/go/lib/arksdk/target/build")
	ignored := patternMatcher.Match(buildDir, true)
	require.Equal(t, false, ignored, "arksdk build directory should not be ignored")
}
