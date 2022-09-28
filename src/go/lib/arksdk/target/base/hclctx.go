package base

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"

	"github.com/myfintech/ark/src/go/lib/ark/kv"

	"github.com/myfintech/ark/src/go/lib/log"

	"github.com/mitchellh/go-homedir"

	"github.com/myfintech/ark/src/go/lib/exec"
	"github.com/pkg/errors"

	"github.com/zclconf/go-cty/cty/function"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"

	"github.com/myfintech/ark/src/go/lib/hclutils"
	"github.com/myfintech/ark/src/go/lib/utils"
)

// EvalContextOptions a set of dependencies for a pre build evaluation context
type EvalContextOptions struct {
	CurrentTarget                Target
	Package                      Package
	TargetLookupTable            LookupTable
	Workspace                    Workspace
	DisableLookupTableEvaluation bool
}

// CreateEvalContext creates a new evaluation context intended for use during Buildable.PreBuild
func CreateEvalContext(opts EvalContextOptions) *hcl.EvalContext {
	ctx := &hcl.EvalContext{
		Functions: hclutils.BuildStdLibFunctions(opts.Workspace.Dir),
		Variables: map[string]cty.Value{
			"workspace": cty.ObjectVal(opts.Workspace.AttributesToCty()),
			"runtime":   hclutils.MapStringStringToCtyObject(utils.GetRuntime()),
			"env":       hclutils.MapStringStringToCtyObject(utils.EnvLookupAll()),
			"package": hclutils.MapStringStringToCtyObject(map[string]string{
				"name":    opts.Package.Name,
				"version": opts.Package.Version,
				"path":    opts.Package.Dir,
			}),
			"cli": hclutils.MapStringStringToCtyObject(map[string]string{
				"args": strings.Join(opts.Workspace.PassableArgs, " "),
				"env":  opts.Workspace.ConfigurationEnvironment,
			}),
		},
	}

	ctx.Functions["kvget"] = KVGet(opts.Workspace.KVStorage)

	if opts.DisableLookupTableEvaluation == false {
		ctx.Variables = hclutils.MergeMapStringCtyValue(opts.TargetLookupTable.ToCtyVariables(),
			ctx.Variables,
		)
	}

	for _, plugin := range opts.Workspace.Config.Plugins {
		ctx.Functions[plugin.Name] = DockerRunFunc(opts.Workspace.Context, plugin)
	}

	if opts.CurrentTarget != nil {
		if buildable, ok := opts.CurrentTarget.(Buildable); ok {
			ctx.Functions["artifact"] = hclutils.PathRelTo(buildable.ArtifactsDir())
			ctx.Functions["load"] = LoadConfigFunc(buildable.PackageDir())
		}
		ctx.Variables["locals"] = cty.ObjectVal(opts.CurrentTarget.LocalVars())

		rawTarget := opts.CurrentTarget.GetRawTarget()
		if rawTarget != nil && rawTarget.Module != nil {
			attributes, _ := rawTarget.Module.EvalAttributes(ctx)
			// FIXME Handle this error
			ctx.Variables["module"] = cty.ObjectVal(map[string]cty.Value{
				"vars": cty.ObjectVal(attributes),
				"path": cty.StringVal(rawTarget.Module.Path),
			})
		}
	}

	return ctx
}

// KVGet pulls data from a local secret store and returns an object
func KVGet(kvStorage kv.Storage) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "path",
				Type: cty.String,
			},
		},
		Type: func(args []cty.Value) (cty.Type, error) {
			kvPath := args[0]

			data, err := getKVData(kvPath.AsString(), kvStorage)
			if err != nil {
				return cty.NilType, errors.Errorf("there was a problem pulling the secret data from the KV store: %v", err)
			}

			jsonBytes, err := json.Marshal(data)
			if err != nil {
				return cty.NilType, err
			}

			return ctyjson.ImpliedType(jsonBytes)
		},
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			kvPath := args[0]

			data, err := getKVData(kvPath.AsString(), kvStorage)
			if err != nil {
				return cty.NullVal(cty.NilType), errors.Errorf("there was a problem pulling the secret data from the KV store: %v", err)
			}

			jsonBytes, err := json.Marshal(data)
			if err != nil {
				return cty.NullVal(cty.NilType), err
			}

			return ctyjson.Unmarshal(jsonBytes, retType)
		},
	})
}

func getKVData(kvPath string, storageConfig kv.Storage) (map[string]interface{}, error) {
	matcher, err := glob.Compile("*.ark/kv*")
	if err != nil {
		return nil, err
	}

	includesBase := matcher.Match(kvPath)
	if includesBase {
		return nil, errors.New("the provided path should not have the base path of the KV store included")
	}

	return storageConfig.Get(kvPath)
}

// DockerRunFunc executes a docker exec using a specifically provided image
func DockerRunFunc(ctx context.Context, plugin Plugin) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name:             "val",
				Type:             cty.DynamicPseudoType,
				AllowDynamicType: true,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			val := args[0]
			if !val.IsWhollyKnown() {
				return cty.UnknownVal(retType), nil
			}

			buf, err := ctyjson.Marshal(val, val.Type())
			if err != nil {
				return cty.NilVal, err
			}

			input := bytes.NewReader(buf)
			output := bytes.NewBuffer(nil)
			if err = exec.DockerExecutor(ctx, exec.DockerExecOptions{
				Image:       plugin.Image,
				Stdin:       input,
				Stdout:      output,
				Stderr:      os.Stderr,
				AttachStdIn: true,
			}); err != nil {
				return cty.NilVal, errors.Wrapf(err, "failed to execute ark plugin: %s\n", plugin.Name)
			}
			if output.String() == "" {
				return cty.NilVal, errors.New("no bytes received on output buffer; this is likely a bug in the plugin container image")
			}
			return cty.StringVal(output.String()), nil
		},
	})
}

// LoadConfigFunc loads environment-specific configuration into the HCL context
func LoadConfigFunc(baseDir string) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "config file path",
				Type: cty.String,
			},
			{
				Name:             "variables",
				Type:             cty.DynamicPseudoType,
				AllowDynamicType: true,
			},
			{
				Name: "disable error handling for invalid file",
				Type: cty.Bool,
			},
		},
		Type: function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			path := args[0].AsString()
			loadVars := args[1].AsValueMap()
			disableErrHandling := args[2].True() // Ref: https://github.com/zclconf/go-cty/blob/main/docs/types.md

			path, err := homedir.Expand(path)
			if err != nil {
				return cty.NilVal, fmt.Errorf("failed to expand ~: %s", err)
			}

			if !filepath.IsAbs(path) {
				path = filepath.Join(baseDir, path)
			}

			// Ensure that the path is canonical for the host OS
			path = filepath.Clean(path)

			baseMap := make(map[string]cty.Value)

			rawHCL, diags := hclutils.FileFromPath(path)
			if diags != nil && diags.HasErrors() && !disableErrHandling {
				log.Info("I made it here")
				return cty.ObjectVal(baseMap), diags
			}

			hclCtx := &hcl.EvalContext{
				Variables: loadVars,
			}

			// if error handling has been disabled, then rawHCL will be nil, which should be alright
			if rawHCL == nil && disableErrHandling {
				return cty.ObjectVal(baseMap), nil
			}

			attrs := hcl.Attributes{}
			if rawHCL != nil {
				attrs, diags = rawHCL.Body.JustAttributes()
				if diags != nil && diags.HasErrors() {
					return cty.ObjectVal(baseMap), diags
				}
			}

			for _, attr := range attrs {
				val, valDiags := attr.Expr.Value(hclCtx)
				if valDiags != nil && valDiags.HasErrors() {
					return cty.ObjectVal(baseMap), valDiags
				}
				baseMap[attr.Name] = val
			}

			return cty.ObjectVal(baseMap), nil
		},
	})
}
