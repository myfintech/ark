package typescript

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/dop251/goja"
	"github.com/pkg/errors"
)

//go:embed embeds/*
var embeddedFS embed.FS

// CompilerOptions an object that represents compilerOptions from tsconfig.json
type CompilerOptions map[string]interface{}

// DefaultCompilerOptions a set of default compiler options for typescript
var DefaultCompilerOptions = CompilerOptions{
	/* Visit https://aka.ms/tsconfig.json to read more about this file */
	"target": "ES2015",   /* Specify ECMAScript target version: 'ES3' (default), 'ES5', 'ES2015', 'ES2016', 'ES2017', 'ES2018', 'ES2019', 'ES2020', or 'ESNEXT'. */
	"module": "commonjs", /* Specify module code generation: 'none', 'commonjs', 'amd', 'system', 'umd', 'es2015', 'es2020', or 'ESNext'. */
	// "lib":                []string{}, /* Specify library files to be included in the compilation. */
	"allowJs":            false, /* Allow javascript files to be compiled. */
	"checkJs":            true,  /* Report errors in .js files. */
	"declaration":        false, /* Generates corresponding '.d.ts' file. */
	"declarationMap":     false, /* Generates a sourcemap for each corresponding '.d.ts' file. */
	"sourceMap":          false, /* Generates corresponding '.map' file. */
	"removeComments":     true,  /* Do not emit comments to output. */
	"noEmit":             true,  /* Do not emit outputs. */
	"importHelpers":      false, /* Import emit helpers from 'tslib'. */
	"downlevelIteration": false, /* Provide full support for iterables in 'for-of', spread, and destructuring when targeting 'ES5' or 'ES3'. */
	"isolatedModules":    true,  /* Transpile each file as a separate module (similar to 'ts.transpileModule'). */

	/* Strict Type-Checking Options */
	"strict":                       true, /* Enable all strict type-checking options. */
	"noImplicitAny":                true, /* Raise error on expressions and declarations with an implied 'any' type. */
	"strictNullChecks":             true, /* Enable strict null checks. */
	"strictFunctionTypes":          true, /* Enable strict checking of function types. */
	"strictBindCallApply":          true, /* Enable strict 'bind', 'call', and 'apply' methods on functions. */
	"strictPropertyInitialization": true, /* Enable strict checking of property initialization in classes. */
	"noImplicitThis":               true, /* Raise error on 'this' expressions with an implied 'any' type. */
	"alwaysStrict":                 true, /* Parse in strict mode and emit "use strict" for each source file. */

	/* Additional Checks */
	"noUnusedLocals":                     true, /* Report errors on unused locals. */
	"noImplicitReturns":                  true, /* Report error when not all code paths in function return a value. */
	"noFallthroughCasesInSwitch":         true, /* Report errors for fallthrough cases in switch statement. */
	"noUncheckedIndexedAccess":           true, /* Include 'undefined' in index signature results */
	"noPropertyAccessFromIndexSignature": true, /* Require undeclared properties from index signatures to use element accesses. */

	/* Module Resolution Options */
	"moduleResolution": "node", /* Specify module resolution strategy: 'node' (Node.js) or 'classic' (TypeScript pre-1.6). */
	"esModuleInterop":  true,   /* Enables emit interoperability between CommonJS and ES Modules via creation of namespace objects for all imports. Implies 'allowSyntheticDefaultImports'. */

	/* Source maps */
	"inlineSourceMap": false, /* Emit a single file with source maps instead of having a separate file. */
	"inlineSources":   false, /* Emit the source alongside the sourcemaps within a single file; requires '--inlineSourceMap' or '--sourceMap' to be set. */

	/* Experimental Options */
	"experimentalDecorators": true, /* Enables experimental support for ES7 decorators. */
	"emitDecoratorMetadata":  true, /* Enables experimental support for emitting type metadata for decorators. */
}

// LoadCompiler loads the embedded version of typescript and compiles it with goja
func LoadCompiler() (*goja.Program, error) {
	filename := "embeds/v4.2.3.js"
	sourceBytes, err := embeddedFS.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return goja.Compile(filename, string(sourceBytes), true)
}

// Transpiler is a wrapper around goja and uses typescript to transpile code
type Transpiler struct {
	typescript      *goja.Program
	runtime         *goja.Runtime
	CompilerOptions CompilerOptions
}

// Transpile uses the provides source and compiler options to transpile from typescript to javascript
func (t *Transpiler) Transpile(source io.Reader) (string, error) {
	_, err := t.runtime.RunProgram(t.typescript)
	if err != nil {
		return "", fmt.Errorf("running typescript compiler: %w", err)
	}

	optionBytes, err := json.Marshal(t.CompilerOptions)
	if err != nil {
		return "", fmt.Errorf("marshalling compile options: %w", err)
	}

	sourceBytes, err := io.ReadAll(source)
	if err != nil {
		return "", errors.Wrap(err, "failed to read typescript source")
	}

	value, err := t.runtime.RunString(fmt.Sprintf("ts.transpile(__go_base64_decode('%s'), %s, /*fileName*/ undefined, /*diagnostics*/ undefined, /*moduleName*/ \"myModule\")",
		base64.StdEncoding.EncodeToString(sourceBytes), optionBytes))

	if err != nil {
		return "", fmt.Errorf("running compiler: %w", err)
	}

	return value.String(), nil
}

func NewTranspiler() (*Transpiler, error) {
	program, err := LoadCompiler()
	if err != nil {
		return nil, err
	}

	return &Transpiler{
		typescript: program,
	}, nil
}

func (t *Transpiler) Install(runtime *goja.Runtime) error {
	t.runtime = runtime
	return runtime.Set("__go_base64_decode", func(call goja.FunctionCall) goja.Value {
		bs, decodeErr := base64.StdEncoding.DecodeString(call.Argument(0).String())
		if decodeErr != nil {
			panic(runtime.ToValue(decodeErr.Error()))
		}
		return runtime.ToValue(string(bs))
	})
}
