package stdlib

import (
	"path/filepath"

	"github.com/myfintech/ark/src/go/lib/watchman"
	"github.com/myfintech/ark/src/go/lib/watchman/wexp"

	"github.com/dop251/goja"
	"github.com/pkg/errors"
	"gopkg.in/osteele/liquid.v1"

	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript/runtime/helpers"
	"github.com/myfintech/ark/src/go/lib/fs"
)

// NewFilepathLibrary creates a module of goja functions for the filepath library
func NewFilepathLibrary(opts Options) typescript.Module {
	return typescript.Module{
		"glob":                             GlobFunc(opts),
		"join":                             JoinFunc(opts),
		"load":                             LoadFunc(opts),
		"fromRoot":                         FromRoot(opts),
		"loadAsTemplate":                   LoadAsTemplateFunc(opts),
		"getFolderNameFromCurrentLocation": GetFolderNameFromCurrentLocationFunc(opts),
	}
}

// GlobFunc returns a goja compatible function that accepts N glob patterns and returns a list of matching files
func GlobFunc(opts Options) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		modPath := helpers.GetCurrentModulePath(opts.Runtime)

		var patterns []string
		for _, argument := range call.Arguments {
			pattern, err := fs.NormalizePathByPrefix(argument.String(), opts.FSRealm, modPath)
			if err != nil {
				panic(opts.Runtime.NewGoError(err))
			}
			patterns = append(patterns, fs.TrimPrefix(pattern, opts.FSRealm))
		}

		var files []string
		queryResp, err := opts.WatchmanClient.Query(watchman.QueryOptions{
			Directory: opts.FSRealm,
			Filter: &watchman.QueryFilter{
				Fields:              watchman.BasicFields(),
				Expression:          wexp.Not(wexp.Type("d")),
				Glob:                patterns,
				DeferVcs:            true,
				DedupResults:        true,
				GlobIncludeDotFiles: true,
			},
		})

		if err != nil {
			panic(opts.Runtime.NewGoError(err))
		}

		for _, file := range queryResp.Files {
			files = append(files, filepath.Join(opts.FSRealm, file.Name))
		}

		return opts.Runtime.ToValue(files)
	}
}

// FromRoot returns an absolute path of the given file from the root
func FromRoot(opts Options) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		path := call.Argument(0).String()
		normalPath, err := fs.NormalizePath(opts.FSRealm, path)
		if err != nil {
			panic(opts.Runtime.NewGoError(err))
		}

		return opts.Runtime.ToValue(normalPath)
	}
}

// JoinFunc joins any number of path elements into a single path,
// separating them with an OS specific Separator. Empty elements
// are ignored. The result is Cleaned. However, if the argument
// list is empty or all its elements are empty, Join returns
// an empty string.
func JoinFunc(opts Options) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		var segments []string
		for _, argument := range call.Arguments {
			segments = append(segments, argument.String())
		}
		return opts.Runtime.ToValue(filepath.Join(segments...))
	}
}

// LoadFunc returns a goja compatible function that loads the contents of a file as a string
func LoadFunc(opts Options) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		modPath := helpers.GetCurrentModulePath(opts.Runtime)
		path, err := fs.NormalizePath(modPath, call.Argument(0).String())
		if err != nil {
			panic(opts.Runtime.NewGoError(err))
		}

		contents, err := fs.ReadFileString(path)
		if err != nil {
			panic(opts.Runtime.NewGoError(err))
		}

		return opts.Runtime.ToValue(contents)
	}
}

type engineDelimiters struct {
	ObjectLeft  string `json:"objectLeft,omitempty"`
	ObjectRight string `json:"objectRight,omitempty"`
	TagLeft     string `json:"tagLeft,omitempty"`
	TagRight    string `json:"tagRight,omitempty"`
}

func defaultEngineDelimiters() *engineDelimiters {
	return &engineDelimiters{
		ObjectLeft:  "${",
		ObjectRight: "}",
		TagLeft:     "%{",
		TagRight:    "~}",
	}
}

// LoadAsTemplateFunc creates a new goja compatible function that accepts a filepath and renders a template
func LoadAsTemplateFunc(opts Options) func(call goja.FunctionCall) goja.Value {
	return helpers.NewGojaErrHandler(opts.Runtime, func(call goja.FunctionCall) (goja.Value, error) {
		vars := make(map[string]interface{})
		delims := defaultEngineDelimiters()
		modPath := helpers.GetCurrentModulePath(opts.Runtime)
		allowNullVars := call.Argument(2).ToBoolean()

		customDelimiters := call.Argument(3)
		if !goja.IsUndefined(customDelimiters) && !goja.IsNull(customDelimiters) {
			err := opts.Runtime.ExportTo(customDelimiters, delims)
			if err != nil {
				return nil, err
			}
		}

		engine := liquid.NewEngine().Delims(
			delims.ObjectLeft,
			delims.ObjectRight,
			delims.TagLeft,
			delims.TagRight,
		)

		path, err := fs.NormalizePath(modPath, call.Argument(0).String())
		if err != nil {
			return nil, err
		}

		contents, err := fs.ReadFileBytes(path)
		if err != nil {
			return nil, err
		}

		parsedTemplate, err := engine.ParseTemplateLocation(contents, path, 1)
		if err != nil {
			return nil, err
		}

		err = opts.Runtime.ExportTo(call.Argument(1), &vars)
		if err != nil {
			return nil, err
		}

		if !allowNullVars {
			for s, i := range vars {
				if i == nil {
					return nil, errors.Errorf("%s %s cannot be null", path, s)
				}
			}
		}

		rendered, err := parsedTemplate.RenderString(vars)
		if err != nil {
			return nil, err
		}

		return opts.Runtime.ToValue(rendered), nil
	})
}

// GetFolderNameFromCurrentLocationFunc return the folder name of the current file
func GetFolderNameFromCurrentLocationFunc(opts Options) func(_ goja.Object) goja.Value {
	return func(_ goja.Object) goja.Value {
		caller := helpers.GetRootCaller(opts.Runtime)

		return opts.Runtime.ToValue(filepath.Base(filepath.Dir(caller)))
	}
}
