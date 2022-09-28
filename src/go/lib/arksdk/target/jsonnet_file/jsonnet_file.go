package jsonnet_file

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-errors/errors"

	"github.com/google/go-jsonnet"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/hclutils"
	"github.com/myfintech/ark/src/go/lib/jsonnetutils"
)

// Target defines the required and optional attributes for defining a jsonnetutils execution
type Target struct {
	*base.RawTarget `json:"-"`

	File       hcl.Expression `hcl:"file,attr"`
	Variables  hcl.Expression `hcl:"variables,attr"`
	LibraryDir hcl.Expression `hcl:"library_dir,optional"`
	Format     hcl.Expression `hcl:"format,optional"`
}

// ComputedAttrs used to store the computed attributes of a local_exec target
type ComputedAttrs struct {
	File       string   `hcl:"file,attr"`
	Variables  *string  `hcl:"variables,attr"`
	LibraryDir []string `hcl:"library_dir,optional"`
	Format     string   `hcl:"format,optional"`
}

// Attributes return combined rawTarget.Attributes with typedTarget.Attributes.
func (t Target) Attributes() map[string]cty.Value {
	return hclutils.MergeMapStringCtyValue(t.RawTarget.Attributes(), map[string]cty.Value{
		"rendered_file": cty.StringVal(t.RenderedFilePath()),
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

// Build constructs a jsonnet manifest from the information provided in the jsonnet target
func (t Target) Build() error {
	attrs := t.ComputedAttrs()

	if filepath.Ext(attrs.File) != ".jsonnet" {
		return errors.Errorf("%s file must have a .jsonnet extension", t.Name)
	}

	if !filepath.IsAbs(attrs.File) {
		attrs.File = filepath.Clean(filepath.Join(t.Dir, attrs.File))
	}

	if err := t.MkArtifactsDir(); err != nil {
		return err
	}

	vm := jsonnet.MakeVM()
	if attrs.Variables != nil {
		vm.NativeFunction(jsonnetArkContext(*attrs.Variables))
	}

	vm.Importer(&jsonnet.FileImporter{
		JPaths: t.Workspace.JsonnetLibrary(attrs.LibraryDir),
	})

	if attrs.Format == "" {
		attrs.Format = "json"
	}

	if attrs.Format != "json" && attrs.Format != "yaml" {
		return errors.Errorf("%s format must be one of (json|yaml)", t.Name)
	}

	input, err := jsonnetutils.ReadInput(false, attrs.File)
	if err != nil {
		return err
	}

	dest, err := os.Create(t.RenderedFilePath())
	if err != nil {
		return err
	}

	defer func() {
		_ = dest.Close()
	}()

	if attrs.Format == "yaml" {
		outputSlice, evalErr := vm.EvaluateSnippetStream(attrs.File, input)
		if evalErr != nil {
			return evalErr
		}
		if writeErr := jsonnetutils.WriteOutputStream(outputSlice, dest); writeErr != nil {
			return writeErr
		}
		return nil
	}

	output, err := vm.EvaluateSnippet(attrs.File, input)
	if err != nil {
		return err
	}
	_, err = dest.WriteString(output)
	if err != nil {
		return err
	}
	return nil
}

// RenderedFilePath path of ark artifacts directory where the manifest will be written to
func (t Target) RenderedFilePath() string {
	attrs := t.ComputedAttrs()
	baseName := filepath.Base(attrs.File)
	return filepath.Join(t.ArtifactsDir(), strings.Replace(baseName, filepath.Ext(baseName), "."+attrs.Format, -1))
}

func jsonnetArkContext(arkCtx string) *jsonnet.NativeFunction {
	return &jsonnet.NativeFunction{
		Name:   "ark_context",
		Params: nil,
		Func: func(input []interface{}) (interface{}, error) {
			var jsonArkCtx map[string]interface{}
			if err := json.Unmarshal([]byte(arkCtx), &jsonArkCtx); err != nil {
				return nil, err
			}
			return jsonArkCtx, nil
		},
	}
}
