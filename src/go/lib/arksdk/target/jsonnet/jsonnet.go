package jsonnet

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/hclutils"
	"github.com/myfintech/ark/src/go/lib/jsonnetutils"

	"github.com/google/go-jsonnet"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// Target defines the required and optional attributes for defining a jsonnetutils execution
type Target struct {
	*base.RawTarget `json:"-"`

	YamlOut    *bool          `hcl:"yaml,attr"`
	Files      hcl.Expression `hcl:"files,attr"`
	OutputDir  hcl.Expression `hcl:"output_dir,attr"`
	Variables  hcl.Expression `hcl:"variables,attr"`
	LibraryDir hcl.Expression `hcl:"library_dir,optional"`
}

// ComputedAttrs used to store the computed attributes of a local_exec target
type ComputedAttrs struct {
	Files      []string `hcl:"files,attr"`
	OutputDir  string   `hcl:"output_dir,attr"`
	Variables  *string  `hcl:"variables,attr"`
	LibraryDir []string `hcl:"library_dir,optional"`
}

// Attributes return combined rawTarget.Attributes with typedTarget.Attributes.
func (t Target) Attributes() map[string]cty.Value {
	return hclutils.MergeMapStringCtyValue(t.RawTarget.Attributes(), map[string]cty.Value{
		// TODO: Review new attributes for this target
		// "output_file": cty.StringVal(t.ConstructOutFilePath()),
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

	if err := os.MkdirAll(attrs.OutputDir, 0755); err != nil {
		return err
	}

	vm := jsonnet.MakeVM()
	if attrs.Variables != nil {
		vm.NativeFunction(jsonnetArkContext(*attrs.Variables))
	}

	vm.Importer(&jsonnet.FileImporter{
		JPaths: t.Workspace.JsonnetLibrary(attrs.LibraryDir),
	})

	for _, file := range attrs.Files {
		input, err := jsonnetutils.ReadInput(false, file)
		if err != nil {
			return err
		}

		dest, err := os.Create(t.ConstructOutFilePath(file))
		if err != nil {
			return err
		}

		defer func() {
			_ = dest.Close()
		}()

		if t.YamlOut != nil && *t.YamlOut {
			outputSlice, evalErr := vm.EvaluateSnippetStream(file, input)
			if evalErr != nil {
				return evalErr
			}
			if writeErr := jsonnetutils.WriteOutputStream(outputSlice, dest); writeErr != nil {
				return writeErr
			}
			continue
		}

		output, err := vm.EvaluateSnippet(file, input)
		if err != nil {
			return err
		}
		_, err = dest.WriteString(output)
		if err != nil {
			return err
		}
	}
	return nil
}

// ConstructOutFilePath builds the correct file path with the proper extension based on the HCL block inputs
func (t Target) ConstructOutFilePath(path string) string {
	attrs := t.ComputedAttrs()
	entryFile := filepath.Base(path)
	outFileBase := strings.TrimSuffix(entryFile, filepath.Ext(entryFile))
	extension := ".json"
	if t.YamlOut != nil && *t.YamlOut {
		extension = ".yaml"
	}
	outFileComplete := fmt.Sprintf("%s%s", outFileBase, extension)
	return filepath.Join(attrs.OutputDir, outFileComplete)
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
