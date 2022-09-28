package hclutils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/stretchr/testify/require"
)

var fooDiff = `--- old/testdata/foo.hcl
+++ new/testdata/foo.hcl
@@ -1,4 +1,4 @@
-thing{
-this=that
-  pianos=strings
+thing {
+  this   = that
+  pianos = strings
 }
`

var dummyStdIn = `foo {
  bar=baz
bat=ban
}
`

var malformedHCL = `foo {
bar=baz
bat=ban
`

func TestFmt(t *testing.T) {
	t.Run("diffing", func(t *testing.T) {
		path := "testdata/foo.hcl"
		file, err := os.Open(path)
		defer func() {
			_ = file.Close()
		}()
		require.NoError(t, err)
		src, err := ioutil.ReadAll(file)
		require.NoError(t, err)
		result := hclwrite.Format(src)
		data, err := bytesDiff(src, result, path)
		require.NoError(t, err)
		require.Equal(t, fooDiff, string(data))
	})
	t.Run("processFile() good file no write", func(t *testing.T) {
		path := "testdata/foo.hcl"
		file, err := os.Open(path)
		defer func() {
			_ = file.Close()
		}()
		require.NoError(t, err)

		opts := &FormatOpts{
			List:      true,
			Write:     false,
			Diff:      false,
			Check:     true,
			Recursive: false,
			Input:     nil,
		}
		if diags := opts.processFile(path, file, os.Stdout); diags != nil && diags.HasErrors() {
			err = diags
		}
		require.NoError(t, err)
	})
	t.Run("processFile() bad file no write", func(t *testing.T) {
		path := "testdata/bar.hcl"
		err := error(nil)
		opts := &FormatOpts{
			List:      true,
			Write:     false,
			Diff:      false,
			Check:     true,
			Recursive: false,
			Input:     nil,
		}
		if diags := opts.processFile(path, os.Stdin, os.Stdout); diags != nil && diags.HasErrors() {
			err = diags
		}
		require.EqualError(t, err, "<nil>: Could not get OS info for provided path; Attempted to get OS information for provided path, but failed to do so.")
	})
	t.Run("processDir() no write", func(t *testing.T) {
		err := error(nil)
		path := "testdata"
		opts := &FormatOpts{
			List:      true,
			Write:     false,
			Diff:      false,
			Check:     true,
			Recursive: false,
			Input:     nil,
		}
		if diags := opts.processDir(path, nil, os.Stdout); diags != nil && diags.HasErrors() {
			err = diags
		}
		require.NoError(t, err)
	})
	t.Run("processDir() no write recursive", func(t *testing.T) {
		err := error(nil)
		path := "testdata"
		opts := &FormatOpts{
			List:      true,
			Write:     false,
			Diff:      false,
			Check:     true,
			Recursive: true,
			Input:     nil,
		}
		if diags := opts.processDir(path, nil, os.Stdout); diags != nil && diags.HasErrors() {
			err = diags
		}
		require.NoError(t, err)
	})
	t.Run("processDir() no write bad path", func(t *testing.T) {
		err := error(nil)
		path := "testdata/dummy"
		opts := &FormatOpts{
			List:      true,
			Write:     false,
			Diff:      false,
			Check:     true,
			Recursive: false,
			Input:     nil,
		}
		if diags := opts.processDir(path, nil, os.Stdout); diags != nil && diags.HasErrors() {
			err = diags
		}
		require.EqualError(t, err, "<nil>: Directory not found; There is no directory at testdata/dummy, and 1 other diagnostic(s)")
	})
	t.Run("fmt() stdin with write false", func(t *testing.T) {
		err := error(nil)
		reader := strings.NewReader(dummyStdIn)
		opts := &FormatOpts{
			List:      true,
			Write:     false,
			Diff:      false,
			Check:     true,
			Recursive: false,
			Input:     nil,
		}
		if diags := opts.fmt(stdinArg, nil, reader, os.Stdout); diags != nil && diags.HasErrors() {
			err = diags
		}
		require.NoError(t, err)
	})
	t.Run("fmt() stdin with write true", func(t *testing.T) {
		err := error(nil)
		opts := &FormatOpts{
			List:      true,
			Write:     true,
			Diff:      false,
			Check:     true,
			Recursive: false,
			Input:     nil,
		}
		if diags := opts.fmt(stdinArg, nil, os.Stdin, os.Stdout); diags != nil && diags.HasErrors() {
			err = diags
		}
		require.EqualError(t, err, "<nil>: Incorrect options passed to fmt command; Option --write cannot be used when reading from stdin.")
	})
	t.Run("fmt() good file with write false", func(t *testing.T) {
		path := "testdata/foo.hcl"
		err := error(nil)
		opts := &FormatOpts{
			List:      true,
			Write:     false,
			Diff:      false,
			Check:     true,
			Recursive: false,
			Input:     nil,
		}
		if diags := opts.fmt(path, nil, os.Stdin, os.Stdout); diags != nil && diags.HasErrors() {
			err = diags
		}
		require.NoError(t, err)
	})
	t.Run("fmt bad file with write false", func(t *testing.T) {
		path := "testdata/bar.hcl"
		err := error(nil)
		opts := &FormatOpts{
			List:      true,
			Write:     false,
			Diff:      false,
			Check:     true,
			Recursive: false,
			Input:     nil,
		}
		if diags := opts.fmt(path, nil, os.Stdin, os.Stdout); diags != nil && diags.HasErrors() {
			err = diags
		}
		require.Error(t, err)
	})
	t.Run("fmt() good dir with write false", func(t *testing.T) {
		path := "testdata"
		err := error(nil)
		opts := &FormatOpts{
			List:      true,
			Write:     false,
			Diff:      false,
			Check:     true,
			Recursive: false,
			Input:     nil,
		}
		if diags := opts.fmt(path, nil, os.Stdin, os.Stdout); diags != nil && diags.HasErrors() {
			err = diags
		}
		require.NoError(t, err)
	})
	t.Run("fmt non-hcl file with write false", func(t *testing.T) {
		path := "testdata/foo.txt"
		err := error(nil)
		opts := &FormatOpts{
			List:      true,
			Write:     false,
			Diff:      false,
			Check:     true,
			Recursive: false,
			Input:     nil,
		}
		if diags := opts.fmt(path, nil, os.Stdin, os.Stdout); diags != nil && diags.HasErrors() {
			err = diags
		}
		require.EqualError(t, err, "<nil>: Non HCL file provided; Only files with the '.hcl' file extension are eligible for formatting.")
	})
	t.Run("fmt() malformedHCL no write", func(t *testing.T) {
		tmpFile := filepath.Join(os.TempDir(), "tmp.hcl")
		err := ioutil.WriteFile(tmpFile, []byte(malformedHCL), 0644)
		require.NoError(t, err)
		opts := &FormatOpts{
			List:      false,
			Write:     false,
			Diff:      false,
			Check:     true,
			Recursive: false,
			Input:     nil,
		}
		if diags := opts.fmt(tmpFile, nil, os.Stdin, os.Stdout); diags != nil && diags.HasErrors() {
			err = diags
		}
		require.Error(t, err)
	})
	t.Run("Run() good file write false check true", func(t *testing.T) {
		path := "testdata/foo.hcl"
		err := error(nil)
		opts := &FormatOpts{
			List:      true,
			Write:     false,
			Diff:      false,
			Check:     true,
			Recursive: false,
			Input:     nil,
		}
		if diags := opts.Run(path, nil); diags != nil && diags.HasErrors() {
			err = diags
		}
		require.EqualError(t, err, "<nil>: Input needs to be formatted; The input is not properly formatted.")
	})
	t.Run("Run() good file write false check false", func(t *testing.T) {
		path := "testdata/foo.hcl"
		err := error(nil)
		opts := &FormatOpts{
			List:      true,
			Write:     false,
			Diff:      false,
			Check:     false,
			Recursive: false,
			Input:     nil,
		}
		if diags := opts.Run(path, nil); diags != nil && diags.HasErrors() {
			err = diags
		}
		require.NoError(t, err)
	})
	t.Run("Run() stdin list true write true", func(t *testing.T) {
		path := stdinArg
		err := error(nil)
		opts := &FormatOpts{
			List:      true,
			Write:     true,
			Diff:      false,
			Check:     false,
			Recursive: false,
			Input:     strings.NewReader(dummyStdIn),
		}
		if diags := opts.Run(path, nil); diags != nil && diags.HasErrors() {
			err = diags
		}
		require.NoError(t, err)
	})
}
