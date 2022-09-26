package helpers

import (
	"path"

	"github.com/dop251/goja"
)

// GetCurrentModulePath evaluates the runtime callstack to determine the current modules path
func GetCurrentModulePath(runtime *goja.Runtime) string {

	return path.Dir(GetCurrentModuleFilePath(runtime))
}

// GetCurrentModuleFilePath evaluates the runtime callstack to determine the current file path
func GetCurrentModuleFilePath(runtime *goja.Runtime) string {
	var buf [2]goja.StackFrame
	frames := runtime.CaptureCallStack(2, buf[:0])
	if len(frames) < 2 {
		return "."
	}
	return frames[1].SrcName()
}

// GetRootCaller evaluates the runtime callstack to determine the root caller
func GetRootCaller(runtime *goja.Runtime) string {
	var buf [2]goja.StackFrame
	frames := runtime.CaptureCallStack(3, buf[:0])
	frameLen := len(frames)
	if frameLen < 2 {
		return "."
	}
	return frames[frameLen-1].SrcName()
}

// NewGojaErrHandler wraps a go function that returns an error and propagates errors to goja's runtime using panic
// This is not idiomatic go but its how goja was designed so instead of writing panics in our goja function we can return errors
// in a more idiomatic way and allow this function to panic for us
func NewGojaErrHandler(runtime *goja.Runtime, f GojaErrFunc) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		val, err := f(call)
		if err != nil {
			panic(runtime.NewGoError(err))
		}
		return val
	}
}

// GojaErrFunc our more idiomatic goja function
type GojaErrFunc func(call goja.FunctionCall) (goja.Value, error)
