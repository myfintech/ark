package hclutils

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/ext/tryfunc"
	"github.com/hashicorp/terraform/lang/funcs"
	"github.com/mitchellh/go-homedir"
	ctyyaml "github.com/zclconf/go-cty-yaml"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

// BuildStdLibFunctions returns a combined map of stdlib functions and feeds scope into filesystem functions
func BuildStdLibFunctions(baseDir string) map[string]function.Function {
	ctxFuncs := map[string]function.Function{
		"abs":              stdlib.AbsoluteFunc,
		"abspath":          funcs.AbsPathFunc,
		"basename":         funcs.BasenameFunc,
		"base64decode":     funcs.Base64DecodeFunc,
		"base64encode":     funcs.Base64EncodeFunc,
		"base64gzip":       funcs.Base64GzipFunc,
		"base64sha256":     funcs.Base64Sha256Func,
		"base64sha512":     funcs.Base64Sha512Func,
		"ceil":             funcs.CeilFunc,
		"chomp":            funcs.ChompFunc,
		"cidrhost":         funcs.CidrHostFunc,
		"cidrnetmask":      funcs.CidrNetmaskFunc,
		"cidrsubnet":       funcs.CidrSubnetFunc,
		"cidrsubnets":      funcs.CidrSubnetsFunc,
		"coalesce":         funcs.CoalesceFunc,
		"coalescelist":     funcs.CoalesceListFunc,
		"compact":          funcs.CompactFunc,
		"concat":           stdlib.ConcatFunc,
		"contains":         funcs.ContainsFunc,
		"csvdecode":        stdlib.CSVDecodeFunc,
		"deepmerge":        DeepMergeFunc(),
		"dirname":          funcs.DirnameFunc,
		"distinct":         funcs.DistinctFunc,
		"element":          funcs.ElementFunc,
		"chunklist":        funcs.ChunklistFunc,
		"file":             funcs.MakeFileFunc(baseDir, false),
		"fileexists":       funcs.MakeFileExistsFunc(baseDir),
		"fileset":          funcs.MakeFileSetFunc(baseDir),
		"filebase64":       funcs.MakeFileFunc(baseDir, true),
		"filebase64sha256": funcs.MakeFileBase64Sha256Func(baseDir),
		"filebase64sha512": funcs.MakeFileBase64Sha512Func(baseDir),
		"filemd5":          funcs.MakeFileMd5Func(baseDir),
		"filesha1":         funcs.MakeFileSha1Func(baseDir),
		"filesha256":       funcs.MakeFileSha256Func(baseDir),
		"filesha512":       funcs.MakeFileSha512Func(baseDir),
		"flatten":          funcs.FlattenFunc,
		"floor":            funcs.FloorFunc,
		"format":           stdlib.FormatFunc,
		"formatdate":       stdlib.FormatDateFunc,
		"formatlist":       stdlib.FormatListFunc,
		"glob":             Glob,
		"indent":           funcs.IndentFunc,
		"index":            funcs.IndexFunc,
		"join":             funcs.JoinFunc,
		"jsondecode":       stdlib.JSONDecodeFunc,
		"jsonencode":       stdlib.JSONEncodeFunc,
		"keys":             funcs.KeysFunc,
		"length":           funcs.LengthFunc,
		"log":              funcs.LogFunc,
		"lookup":           funcs.LookupFunc,
		"lower":            stdlib.LowerFunc,
		"matchkeys":        funcs.MatchkeysFunc,
		"max":              stdlib.MaxFunc,
		"md5":              funcs.Md5Func,
		"merge":            funcs.MergeFunc,
		"min":              stdlib.MinFunc,
		"parseint":         funcs.ParseIntFunc,
		"pathexpand":       funcs.PathExpandFunc,
		"pow":              funcs.PowFunc,
		"range":            stdlib.RangeFunc,
		"regex":            stdlib.RegexFunc,
		"regexall":         stdlib.RegexAllFunc,
		"replace":          funcs.ReplaceFunc,
		"reverse":          funcs.ReverseFunc,
		"rsadecrypt":       funcs.RsaDecryptFunc,
		"setintersection":  stdlib.SetIntersectionFunc,
		"setproduct":       funcs.SetProductFunc,
		"setsubtract":      stdlib.SetSubtractFunc,
		"setunion":         stdlib.SetUnionFunc,
		"sha1":             funcs.Sha1Func,
		"sha256":           funcs.Sha256Func,
		"sha512":           funcs.Sha512Func,
		"signum":           funcs.SignumFunc,
		"slice":            funcs.SliceFunc,
		"sort":             funcs.SortFunc,
		"split":            funcs.SplitFunc,
		"strrev":           stdlib.ReverseFunc,
		"substr":           stdlib.SubstrFunc,
		"title":            funcs.TitleFunc,
		"tostring":         funcs.MakeToFunc(cty.String),
		"tonumber":         funcs.MakeToFunc(cty.Number),
		"tobool":           funcs.MakeToFunc(cty.Bool),
		"toset":            funcs.MakeToFunc(cty.Set(cty.DynamicPseudoType)),
		"tolist":           funcs.MakeToFunc(cty.List(cty.DynamicPseudoType)),
		"tomap":            funcs.MakeToFunc(cty.Map(cty.DynamicPseudoType)),
		"transpose":        funcs.TransposeFunc,
		"trim":             funcs.TrimFunc,
		"trimprefix":       funcs.TrimPrefixFunc,
		"trimspace":        funcs.TrimSpaceFunc,
		"trimsuffix":       funcs.TrimSuffixFunc,
		"try":              tryfunc.TryFunc,
		"upper":            stdlib.UpperFunc,
		"urlencode":        funcs.URLEncodeFunc,
		"values":           funcs.ValuesFunc,
		"yamldecode":       ctyyaml.YAMLDecodeFunc,
		"yamlencode":       ctyyaml.YAMLEncodeFunc,
		"zipmap":           funcs.ZipmapFunc,
	}

	ctxFuncs["templatefile"] = funcs.MakeTemplateFileFunc(baseDir, func() map[string]function.Function {
		// The templatefile function prevents recursive calls to itself
		// by copying this map and overwriting the "templatefile" entry.
		return ctxFuncs
	})

	return ctxFuncs
}

// PathRelTo constructs a function that takes a dir path, cleans, mkdirs, and returns it
func PathRelTo(baseDir string) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "path",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			path := args[0].AsString()
			path, err := homedir.Expand(path)
			if err != nil {
				return cty.NilVal, fmt.Errorf("failed to expand ~: %s", err)
			}

			if !filepath.IsAbs(path) {
				path = filepath.Join(baseDir, path)
			}

			// Ensure that the path is canonical for the host OS
			path = filepath.Clean(path)

			return cty.StringVal(path), nil
		},
	})
}

var Glob = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "pattern",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.List(cty.String)),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		pattern := args[0].AsString()
		var paths []cty.Value

		glob, err := filepath.Glob(pattern)
		if err != nil {
			return cty.NilVal, fmt.Errorf("failed to build file list from pattern: %s", err)
		}

		for _, i := range glob {
			paths = append(paths, cty.StringVal(i))
		}

		return cty.ListVal(paths), nil
	},
})

// DeepMergeFunc works the same as the regular merge() but is able to do deeply nested data structures
func DeepMergeFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{},
		VarParam: &function.Parameter{
			Name:             "maps",
			Type:             cty.DynamicPseudoType,
			AllowDynamicType: true,
			AllowNull:        true,
		},
		Type: func(args []cty.Value) (cty.Type, error) {
			// empty args is accepted, so assume an empty object since we have no
			// key-value types.
			if len(args) == 0 {
				return cty.EmptyObject, nil
			}

			// check for invalid arguments
			for _, arg := range args {
				ty := arg.Type()

				// we can't work with dynamic types, so move along
				if ty.Equals(cty.DynamicPseudoType) {
					return cty.DynamicPseudoType, nil
				}

				if !ty.IsMapType() && !ty.IsObjectType() {
					return cty.NilType, fmt.Errorf("arguments must be maps or objects, got %#v", ty.FriendlyName())
				}
			}

			// pre-merge objects so we can determine final type
			outputMap := cty.NullVal(cty.NilType)

			for _, arg := range args {
				outputMap = recursiveMerge(arg, outputMap)
			}

			if outputMap.Type().HasDynamicTypes() {
				return cty.DynamicPseudoType, nil
			}

			return outputMap.Type(), nil
		},
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			outputMap := cty.NullVal(cty.NilType)

			for _, arg := range args {
				outputMap = recursiveMerge(arg, outputMap)
			}

			return outputMap, nil
		},
	})
}

func recursiveMerge(newValue cty.Value, existingValue cty.Value) cty.Value {

	// unknown types trump known ones, don't merge
	if !existingValue.IsKnown() {
		return existingValue
	}

	// New value isn't mergeable, so just replace.
	newValueMergeable := newValue.Type().IsMapType() || newValue.Type().IsObjectType()
	existingValueMergeable := existingValue.Type().IsMapType() || existingValue.Type().IsObjectType()
	if !newValueMergeable || !existingValueMergeable || existingValue.IsNull() {
		return newValue
	}

	if newValue.IsNull() {
		return existingValue
	}

	mergedMap := existingValue.AsValueMap()
	for it := newValue.ElementIterator(); it.Next(); {
		k, newValue := it.Element()
		key := k.AsString()
		if newValue.IsNull() {
			delete(mergedMap, key)
			continue
		}

		mergedMap[key] = recursiveMerge(newValue, mergedMap[key])
	}
	// strip out any null properties
	for key, value := range mergedMap {
		if value.IsNull() {
			delete(mergedMap, key)
		}
	}

	return wrapAsObjectOrMap(mergedMap)
}

func wrapAsObjectOrMap(input map[string]cty.Value) cty.Value {
	if len(input) == 0 {
		return cty.EmptyObjectVal
	}

	typesMatch := true
	firstType := cty.NilType

	for _, value := range input {
		ty := value.Type()
		if firstType == cty.NilType {
			firstType = ty
		}

		if !ty.Equals(firstType) {
			typesMatch = false
		}
	}

	if typesMatch {
		return cty.MapVal(input)
	}

	return cty.ObjectVal(input)
}
