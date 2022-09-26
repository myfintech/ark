package stdlib

import (
	"encoding/base64"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/dop251/goja"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript"
)

// EncodingLibraryOptions input type for NewEncodingLibrary factory
type EncodingLibraryOptions struct {
	Runtime *goja.Runtime
}

// NewEncodingLibrary register all TS-Golang mappings
func NewEncodingLibrary(opts EncodingLibraryOptions) typescript.Module {
	return typescript.Module{
		"base64":      base64Func(opts),
		"json2string": json2stringFunc(opts),
	}
}

func base64Func(opts EncodingLibraryOptions) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		runtime := opts.Runtime
		op := call.Argument(1).String()
		value := call.Argument(0).String()

		if op == "undefined" || op == "" {
			// panic(runtime.NewGoError(errors.New("operation cannot be null or empty")))
			op = "encode"
		}

		if value == "" {
			panic(runtime.NewGoError(errors.New("value cannot be null or empty")))
		}

		result := ""
		if op == "encode" {
			result = base64.StdEncoding.EncodeToString([]byte(value))
		}

		if op == "decode" {
			data, err := base64.StdEncoding.DecodeString(value)
			if err != nil {
				panic(runtime.NewGoError(err))
			}

			result = string(data)
		}

		return runtime.ToValue(result)
	}
}

func json2stringFunc(opts EncodingLibraryOptions) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		runtime := opts.Runtime
		value := call.Argument(0).ToObject(runtime)

		result, err := json.Marshal(value)

		if err != nil {
			panic(runtime.NewGoError(err))
		}

		return runtime.ToValue(string(result))
	}
}
