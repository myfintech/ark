package base

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zclconf/go-cty/cty"

	"github.com/myfintech/ark/src/go/lib/hclutils"
)

type exampleTarget struct {
	*RawTarget
	NotARawField string `hcl:"not_a_raw_field,attr"`
	ExtraFields  string `hcl:"extra_fields,attr"`
}

func (t exampleTarget) Build() error                             { return nil }
func (t exampleTarget) PreBuild() error                          { return nil }
func (t exampleTarget) ComputedAttrsToCty() map[string]cty.Value { return make(map[string]cty.Value) }

func TestRawTarget(t *testing.T) {
	cwd, _ := os.Getwd()
	workspace := NewWorkspace()
	testdata := filepath.Join(cwd, "testdata")
	require.NoError(t, workspace.DetermineRoot(testdata))

	workspace.RegisteredTargets = Targets{
		"example": exampleTarget{},
	}

	buildFiles, err := workspace.DecodeBuildFiles()
	require.NoError(t, err)

	err = workspace.LoadTargets(buildFiles)
	require.NoError(t, err)

	t.Run("raw_target hashing should be consistent", func(t *testing.T) {
		target, lookupErr := workspace.TargetLUT.LookupByAddress("test.example.foo")
		require.NoError(t, lookupErr)

		example, ok := target.(exampleTarget)
		require.Equal(t, true, ok, "target should be an exampleTarget type")

		isolatedBlocks, isoErr := example.IsolateHCLBlocks()
		require.NoError(t, isoErr)

		hash, hashErr := hclutils.HashFile(isolatedBlocks, nil)
		require.NoError(t, hashErr)

		// If this test fails in the future it may be because of HashableAttributes returning a map[string]interface{}
		require.Equal(t, "806b107b44986f038b964867bf7bbdb58d1f7b78f9cc606d5616cc00c39051d4", hex.EncodeToString(hash.Sum(nil)), "target attributes should have a consistent hash")
		require.Equal(t, "6963b35535aba66e51fddfc4c62840708028d2cd9ef4953195df8b35bdc95d00", example.Hash(), "target hash should be consistent")

		fmCache, _ := workspace.Observer.GetMatchCache(example.Address())
		var matchFileRelPaths []string
		for _, file := range fmCache.FilesList() {
			matchFileRelPaths = append(matchFileRelPaths, file.RelName)
		}
		require.NotContains(t, matchFileRelPaths, ".gitignore")
		require.NotContains(t, matchFileRelPaths, "ignore_me.txt")
	})

	t.Run("should be able to map example target type", func(t *testing.T) {
		target, lookupErr := workspace.TargetLUT.LookupByAddress("test.example.foo")
		require.NoError(t, lookupErr)

		example, ok := target.(exampleTarget)
		require.Equal(t, true, ok, "target should be an exampleTarget type")
		// require.Equal(t, rawTargets[0], example.RawTarget)
		require.Equal(t, "testing", example.NotARawField)
		require.Empty(t, example.DependsOn)
		require.NotEmpty(t, example.SourceFiles)
	})

	t.Run("should be able to hash declared source files", func(t *testing.T) {
		target, lookupErr := workspace.TargetLUT.LookupByAddress("test.example.foo")
		require.NoError(t, lookupErr)

		example, ok := target.(exampleTarget)
		require.Equal(t, true, ok, "target should be an exampleTarget type")

		t.Log(example.Hash())
	})

	t.Run("should be able to get isolated hcl blocks for hashing", func(t *testing.T) {
		target, lookupErr := workspace.TargetLUT.LookupByAddress("test.example.foo")
		require.NoError(t, lookupErr)

		_, ok := target.(hclutils.Isolator)
		require.Equal(t, true, ok, "target should be an exampleTarget type")
	})

	t.Run("should be able to get labels", func(t *testing.T) {
		target, lookupErr := workspace.TargetLUT.LookupByAddress("test.example.foo")
		require.NoError(t, lookupErr)

		example, ok := target.(exampleTarget)
		require.Equal(t, true, ok, "target should be an exampleTarget type")

		labels := example.ListLabels()
		require.NotEmpty(t, labels, "labels string slice should not be empty")
		require.Equal(t, []string{"test1", "test2", "test3"}, labels)
	})

}
