package stdlib

import (
	"github.com/dop251/goja"
	"github.com/myfintech/ark/src/go/lib/ark/kv"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript"
)

// KvLibraryOptions input type for NewKvLibrary factory
type KvLibraryOptions struct {
	KVStorage kv.Storage
	Runtime   *goja.Runtime
}

// NewKvLibrary register all TS-Golang mappings
func NewKvLibrary(opts KvLibraryOptions) typescript.Module {
	return typescript.Module{
		"get": getFunc(opts),
	}
}

func getFunc(opts KvLibraryOptions) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		kvStorage := opts.KVStorage
		runtime := opts.Runtime
		path := call.Argument(0).String()

		secretData, err := kvStorage.Get(path)
		if err != nil {
			panic(runtime.NewGoError(err))
		}

		return runtime.ToValue(secretData)
	}
}
