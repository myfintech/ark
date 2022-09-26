package hclutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/utils/cryptoutils"
)

// FileFromString uses a new parser and attempts to load HCL from a string
// we generate a hash of the string and use it as the filename as the ParseHCL function requires a filename to avoid parsing the same file
func FileFromString(rawHCL string) (*hcl.File, hcl.Diagnostics) {
	sha, err := cryptoutils.SHA256Sum(rawHCL, "hex")
	if err != nil {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "failed to generate shasum of rawHCL",
				Detail:   err.Error(),
			},
		}
	}
	return hclparse.NewParser().ParseHCL([]byte(rawHCL), sha)
}

// FileFromPath reads data from a file, parses it as HCL and returns a pointer to an HCLFile struct
func FileFromPath(filename string) (*hcl.File, hcl.Diagnostics) {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return &hcl.File{}, hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "Configuration file not found",
					Detail:   fmt.Sprintf("The configuration file %s does not exist.", filename),
				},
			}
		}
		return &hcl.File{}, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Failed to read configuration",
				Detail:   fmt.Sprintf("Can't read %s: %s.", filename, err),
			},
		}
	}
	return hclsyntax.ParseConfig(src, filename, hcl.Pos{Line: 1, Column: 1})
}

// StructField a reflected struct field, and it's parsed tags.
type StructField struct {
	Value     reflect.Value
	Field     reflect.StructField
	Tag       reflect.StructTag
	TagString string
	TagValues []string
}

// StructFields extracts all HCL tagged struct fields and returns a map of reflections by name
// If a tagged field does not have a name (hcl:",label") it is excluded
func StructFields(v interface{}) (map[string]StructField, error) {
	fields := map[string]StructField{}

	rv := reflect.ValueOf(v)
	rt := reflect.TypeOf(v)

	if rv.Kind() != reflect.Ptr {
		return fields, errors.Errorf("%s must be a pointer", rv.Type().String())
	}

	if rv.Elem().Kind() != reflect.Struct {
		return fields, errors.Errorf("%s must be a pointer to a struct", rv.Type().String())
	}

	elem := rt.Elem()
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		if tag, ok := field.Tag.Lookup("hcl"); ok {
			tagValues := strings.Split(tag, ",")
			if len(tagValues) > 0 && tagValues[0] != "" {
				fields[tagValues[0]] = StructField{
					Value:     rv.Elem().Field(i),
					Field:     field,
					Tag:       field.Tag,
					TagString: tag,
					TagValues: tagValues,
				}
			}
		}
	}
	return fields, nil
}

// TraversalToKeys translates an hcl.Traversal to a slice of string keys.
func TraversalToKeys(traversal hcl.Traversal) []string {
	var keys []string
	for _, traversalSegment := range traversal {
		switch val := traversalSegment.(type) {
		case hcl.TraverseRoot:
			keys = append(keys, val.Name)
			continue
		case hcl.TraverseAttr:
			keys = append(keys, val.Name)
			continue
		}
	}
	return keys
}

// TraversalToString the same as TraversalToKeys but returns a string representation
func TraversalToString(traversal hcl.Traversal) string {
	return strings.Join(TraversalToKeys(traversal), ".")
}

// DiagToErrWrap combines all diag errors into a single error value
// If diag.HasErrors() returns false this function returns a nil error
func DiagToErrWrap(diag hcl.Diagnostics) (err error) {
	if diag.HasErrors() {
		var errStrings []string
		for _, diagErr := range diag.Errs() {
			errStrings = append(errStrings, diagErr.Error())
		}
		err = errors.New(strings.Join(errStrings, "\n"))
	}
	return
}

// DecodeExpressions takes a set of struct pointers (src, dest)
// extracts src and dest struct fields by tag of HCL
// checks each hcl field for type of hcl.Expression and decodes the value of that expression into the matching dest struct field.
func DecodeExpressions(src interface{}, dest interface{}, eval *hcl.EvalContext) error {
	srcHCLFields, err := StructFields(src)
	if err != nil {
		return err
	}
	destHCLFields, err := StructFields(dest)
	if err != nil {
		return err
	}

	for srcFieldName, srcField := range srcHCLFields {
		if exp, isExp := srcField.Value.Interface().(hcl.Expression); isExp {
			if destField, match := destHCLFields[srcFieldName]; match {
				if diag := DecodeExpressionFromReflection(exp, destField, eval); diag.HasErrors() {
					return DiagToErrWrap(diag)
				}
			}
		}
	}
	return nil
}

// DecodeExpressionFromReflection takes a reflected value and decodes an hcl expression into it with an eval context.
func DecodeExpressionFromReflection(exp hcl.Expression, field StructField, eval *hcl.EvalContext) hcl.Diagnostics {
	diag := hcl.Diagnostics{}
	switch field.Value.Kind() {
	case reflect.String:
		tempStr := field.Value.String()
		if diag = gohcl.DecodeExpression(exp, eval, &tempStr); !diag.HasErrors() {
			field.Value.SetString(tempStr)
		}
	default:
		diag = gohcl.DecodeExpression(exp, eval, field.Value.Addr().Interface())
	}

	if diag.HasErrors() {
		for _, tv := range field.TagValues {
			// FIXME: checking strings for null may not be reliable enough
			if tv == "optional" && strings.Contains(diag.Error(), "null") {
				return nil
			}
		}
		diag = diag.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to compute attribute",
			Detail:   field.Field.Name,
			Subject:  exp.StartRange().Ptr(),
			Context:  exp.Range().Ptr(),
		})
	}

	return diag
}

// SortedAttrs is a slice of HCL attributes that implements sort.Interface
type SortedAttrs []*hcl.Attribute

// Len returns the length of a data set
func (s SortedAttrs) Len() int {
	return len(s)
}

// Less returns a boolean for which item is less than a comparison item
func (s SortedAttrs) Less(i, j int) bool {
	return len(s[i].Expr.Variables()) < len(s[j].Expr.Variables())
}

// Swap exchanges the values between two indexes
func (s SortedAttrs) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// SortAttributes returns a sorted slice of attributes so variable decoding happens in the correct order
func SortAttributes(attributes hcl.Attributes) SortedAttrs {
	sortedAttrs := make(SortedAttrs, 0)
	for _, v := range attributes {
		sortedAttrs = append(sortedAttrs, v)
	}
	sort.Sort(sortedAttrs)
	return sortedAttrs
}
