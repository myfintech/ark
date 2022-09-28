package local_file

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/myfintech/ark/src/go/lib/fs"

	"github.com/hashicorp/hcl/v2"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/hclutils"
)

const (
	createDirError = "unable to create directory structure"
	writeFileError = "unable to write file"
)

// Target an executable target, when built, it creates a file from provided data
type Target struct {
	*base.RawTarget `json:"-"`

	Filename             hcl.Expression `hcl:"filename,attr"`
	Content              hcl.Expression `hcl:"content,attr"`
	FilePermissions      hcl.Expression `hcl:"file_permissions,optional"`
	DirectoryPermissions hcl.Expression `hcl:"directory_permissions,optional"`
	// Cacheable            hcl.Expression `hcl:"cacheable,optional"` // TODO: add optional cacheable feature for non-dynamic file targets
}

// ComputedAttrs used to store the computed attributes of a local_file target
type ComputedAttrs struct {
	Filename             string      `hcl:"filename,attr"`
	Content              string      `hcl:"content,attr"`
	FilePermissions      os.FileMode `hcl:"file_permissions,optional"`
	DirectoryPermissions os.FileMode `hcl:"directory_permissions,optional"`
	// Cacheable            bool        `hcl:"cacheable,optional"`
}

// Attributes returns a combined map of rawTarget.Attributes and typedTarget.Attributes
func (t Target) Attributes() map[string]cty.Value {
	// computed := t.ComputedAttrs()
	return hclutils.MergeMapStringCtyValue(t.RawTarget.Attributes(), map[string]cty.Value{
		"filename": cty.StringVal(t.Artifact()),
	})
}

// ComputedAttrs returns a pointer to computed attributes from the state store.
// If attributes are not in the state store it will create a new pointer and insert it into the state store.
func (t Target) ComputedAttrs() *ComputedAttrs {
	if attrs, ok := t.GetStateAttrs().(*ComputedAttrs); ok {
		return attrs
	}

	attrs := &ComputedAttrs{}
	t.SetStateAttrs(attrs)
	return attrs
}

// CacheEnabled overrides the default target caching behavior
func (t Target) CacheEnabled() {
	return
}

// PreBuild a lifecycle hook for calculating state before the build
func (t Target) PreBuild() error {
	attrs := t.ComputedAttrs()
	err := hclutils.DecodeExpressions(&t, attrs, base.CreateEvalContext(base.EvalContextOptions{
		CurrentTarget:     t,
		Package:           *t.Package,
		TargetLookupTable: t.Workspace.TargetLUT,
		Workspace:         *t.Workspace,
	}))
	if err != nil {
		return err
	}
	if attrs.FilePermissions == 0 {
		attrs.FilePermissions = 0644
	}
	if attrs.DirectoryPermissions == 0 {
		attrs.DirectoryPermissions = 0755
	}
	return nil
}

// Build creates the file from the content in the location specified with the permissions specified
func (t Target) Build() error {
	attrs := t.ComputedAttrs()
	fileBytes := []byte(attrs.Content)

	if attrs.Content == "" {
		return errors.New("no content present to write to file")
	}

	isAbs := filepath.IsAbs(attrs.Filename)
	if err := t.MkArtifactsDir(); err != nil {
		return errors.Wrap(err, createDirError)
	}

	if writeErr := ioutil.WriteFile(t.Artifact(), fileBytes, attrs.FilePermissions); writeErr != nil {
		return errors.Wrap(writeErr, writeFileError)
	}

	if isAbs {
		return t.CopyArtifact()
	}

	return nil
}

// CheckLocalBuildCache loads the build cache state and verifies that the hashes match
func (t Target) CheckLocalBuildCache() (bool, error) {
	cached, err := t.RawTarget.CheckLocalBuildCache()
	if err != nil {
		return false, err
	}

	if !cached {
		return false, nil
	}

	if preBuildErr := t.PreBuild(); preBuildErr != nil {
		return false, errors.Wrap(preBuildErr, "unable to run prebuild")
	}
	attrs := t.ComputedAttrs()

	if !filepath.IsAbs(attrs.Filename) {
		return true, nil
	}

	_, err = os.Stat(attrs.Filename)
	if os.IsNotExist(err) {
		return true, t.CopyArtifact()
	}
	if err != nil {
		return false, err
	}

	return true, t.CopyArtifact()
}

// CopyArtifact makes a copy of an artifact in a desired location
func (t Target) CopyArtifact() error {
	attrs := t.ComputedAttrs()

	directoryStructure := filepath.Dir(attrs.Filename)
	if mkDirErr := os.MkdirAll(directoryStructure, attrs.DirectoryPermissions); mkDirErr != nil {
		return errors.Wrap(mkDirErr, createDirError)
	}
	err := fs.Copy(t.Artifact(), attrs.Filename)
	if os.IsExist(err) {
		if rmerr := os.Remove(attrs.Filename); rmerr != nil {
			return rmerr
		}
		return fs.Copy(t.Artifact(), attrs.Filename)
	}
	if err != nil {
		return err
	}
	return nil
}

// Artifact returns the absolute path to an artifact
func (t Target) Artifact() string {
	attrs := t.ComputedAttrs()
	return filepath.Clean(filepath.Join(t.ArtifactsDir(), filepath.Base(attrs.Filename)))
}
