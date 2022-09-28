package http_archive

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/fs"
	"github.com/myfintech/ark/src/go/lib/hclutils"
)

var httpPrefix = regexp.MustCompile("^https?://")

// Target an executable target, when built, it creates a file from provided data
type Target struct {
	*base.RawTarget `json:"-"`
	URL             hcl.Expression `hcl:"url,attr"`
	Sha256          hcl.Expression `hcl:"sha256,attr"`
	Decompress      hcl.Expression `hcl:"decompress,attr"`
}

// ComputedAttrs used to store the computed attributes of a local_file target
type ComputedAttrs struct {
	URL        string `hcl:"url,attr"`
	Sha256     string `hcl:"sha256,attr"`
	Decompress bool   `hcl:"decompress,optional"`
}

// Attributes returns a combined map of rawTarget.Attributes and typedTarget.Attributes
func (t Target) Attributes() map[string]cty.Value {
	computed := t.ComputedAttrs()
	return hclutils.MergeMapStringCtyValue(t.RawTarget.Attributes(), map[string]cty.Value{
		"url":           cty.StringVal(computed.URL),
		"contents_path": cty.StringVal(t.ArtifactsDir()),
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

// PreBuild a lifecycle hook for calculating state before the build
func (t Target) PreBuild() error {
	return hclutils.DecodeExpressions(&t, t.ComputedAttrs(), base.CreateEvalContext(base.EvalContextOptions{
		CurrentTarget:     t,
		Package:           *t.Package,
		TargetLookupTable: t.Workspace.TargetLUT,
		Workspace:         *t.Workspace,
	}))
}

// Build creates the file from the content in the location specified with the permissions specified
func (t Target) Build() (err error) {
	attrs := t.ComputedAttrs()

	if !httpPrefix.MatchString(attrs.URL) {
		return errors.Errorf("%s must use a secure URL (https)", t.Name)
	}

	if attrs.Sha256 == "" {
		return errors.Errorf("%s must provide a sha256 checksum for integrity verification", t.Sha256)
	}

	defer func() {
		_ = os.RemoveAll(t.TmpArtifactsDir())
	}()

	if err = t.MkTmpArtifactsDir(); err != nil {
		return
	}

	if err = t.MkArtifactsDir(); err != nil {
		return
	}

	baseName := filepath.Base(attrs.URL)
	artifactPath := filepath.Join(t.ArtifactsDir(), baseName)
	tmpFilePath, err := fs.Download(attrs.URL, t.TmpArtifactsDir(), 0755)
	if err != nil {
		return
	}

	_, err = fs.CompareFileHash(tmpFilePath, attrs.Sha256)
	if err != nil {
		return errors.Wrap(err, "http_archive#target.Build")
	}

	if err = fs.Copy(tmpFilePath, artifactPath); err != nil {
		return
	}

	if attrs.Decompress {
		artifactFile, openErr := os.Open(artifactPath)
		if openErr != nil {
			return openErr
		}
		defer func() {
			_ = artifactFile.Close()
			_ = os.Remove(artifactPath)
		}()

		if err = fs.GzipUntar(t.ArtifactsDir(), artifactFile); err != nil {
			return
		}
	}

	return nil
}
