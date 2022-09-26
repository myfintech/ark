package typescript

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dop251/goja"
	"github.com/pkg/errors"

	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript/runtime/helpers"
)

// Library represents an alias for library paths
type Library struct {
	Path   string
	Prefix string
}

// ModuleResolver a goja runtime plugin that provides module resolution and automatic typescript transpilation
type ModuleResolver struct {
	Transpiler      *Transpiler
	CompilerOptions map[string]interface{}
	Libraries       []Library

	runtime *goja.Runtime
	modules sync.Map
}

// DefaultModuleResolver creates a pointer to a ModuleResolver with the DefaultCompilerOptions
func DefaultModuleResolver(transpiler *Transpiler) *ModuleResolver {
	return &ModuleResolver{
		Transpiler:      transpiler,
		CompilerOptions: DefaultCompilerOptions,
	}
}

// SearchLibraries resolves path on a given library
func (m *ModuleResolver) SearchLibraries(modulePath string) string {
	for _, lib := range m.Libraries {
		if strings.HasPrefix(modulePath, lib.Prefix) {
			return strings.Replace(modulePath, lib.Prefix, lib.Path, -1)
		}
	}

	return ""
}

// Install the module resolver into a Runtime instance
func (m *ModuleResolver) Install(runtime *goja.Runtime) error {
	m.runtime = runtime
	return m.runtime.Set("require", func(call goja.FunctionCall) goja.Value {
		modulePath := call.Argument(0).String()

		// attempt to load a native module first
		if mod := m.get(modulePath); mod != nil {
			return mod.Get("exports")
		}

		if !filepath.IsAbs(modulePath) && !strings.HasPrefix(modulePath, ".") {
			if found := m.SearchLibraries(modulePath); found != "" {
				modulePath = found
			}
		}

		if !filepath.IsAbs(modulePath) {
			modulePath = filepath.Join(helpers.GetCurrentModulePath(m.runtime), modulePath)
		}

		modulePath, err := resolveModuleFile(modulePath)
		if err != nil {
			return m.runtime.NewGoError(err)
		}

		resolver := m.resolveAndTranspile
		// switch ext {
		// case ".ts":
		// 	break
		// default:
		// 	panic(
		// 		m.runtime.NewGoError(
		// 			errors.Errorf("cannot load module %s extension %s not supported",
		// 				modulePath, ext),
		// 		),
		// 	)
		// }

		mod, err := resolver(modulePath)
		if err != nil {
			panic(m.runtime.NewGoError(err))
		}
		return mod.Get("exports")
	})
}

func resolveModuleFile(modpath string) (string, error) {
	stat, err := os.Stat(modpath)

	// mod exists and isn't a directory
	if err == nil && !stat.IsDir() {
		return modpath, nil
	}

	// mod exists and is a directory
	if err == nil && stat.IsDir() {
		// prefer a file over folder/index.ts
		if _, sErr := os.Stat(modpath + ".ts"); sErr == nil {
			return modpath + ".ts", nil
		}

		indexPath := filepath.Join(modpath, "index.ts")
		// mod is contains an index.ts
		if stat, err = os.Stat(indexPath); err == nil {
			return indexPath, nil
		}
	}

	// mod doesn't exist and has a known extension (.ts, .js, .json)
	// nothing more we can try here
	if os.IsNotExist(err) && hashKnownExt(modpath) {
		return modpath, err
	}

	// mod doesn't exist and doesn't have known extension
	// we should try checking ".ts"
	if os.IsNotExist(err) && !hashKnownExt(modpath) {
		modpath += ".ts"
		stat, err = os.Stat(modpath)
	}

	return modpath, err
}

func hashKnownExt(modpath string) bool {
	switch filepath.Ext(modpath) {
	case ".ts", ".js", ".json":
		return true
	default:
		return false
	}
}

func (m *ModuleResolver) set(path string, module *goja.Object) {
	m.modules.Store(path, module)
}

func (m *ModuleResolver) get(path string) *goja.Object {
	if v, ok := m.modules.Load(path); ok {
		return v.(*goja.Object)
	}
	return nil
}

func (m *ModuleResolver) resolveAndTranspile(path string) (module *goja.Object, err error) {
	if module = m.get(path); module != nil {
		return
	}

	file, err := os.Open(path)
	if err != nil {
		return
	}

	transpiledSource, err := m.Transpiler.Transpile(file)
	if err != nil {
		return
	}

	moduleProgram, err := m.wrapAndCompileModule(path, transpiledSource)
	if err != nil {
		return
	}

	module = m.runtime.NewObject()
	if err = module.Set("exports", m.runtime.NewObject()); err != nil {
		return
	}

	if err = m.loadModuleFile(moduleProgram, module); err != nil {
		return
	}
	m.set(path, module)
	return
}

func (m *ModuleResolver) wrapAndCompileModule(filename, source string) (*goja.Program, error) {
	transpiledCode := "(function(exports, require, module) {\n" + source + "\n})"
	return goja.Compile(filename, transpiledCode, true)
}

func (m *ModuleResolver) loadModuleFile(program *goja.Program, jsModule *goja.Object) error {
	f, err := m.runtime.RunProgram(program)
	if err != nil {
		return err
	}

	if call, ok := goja.AssertFunction(f); ok {
		jsExports := jsModule.Get("exports")
		jsRequire := m.runtime.Get("require")

		// Run the module source, with "jsExports" as "this",
		// "jsExports" as the "exports" variable, "jsRequire"
		// as the "require" variable and "jsModule" as the
		// "module" variable (Nodejs capable).
		_, err = call(jsExports, jsExports, jsRequire, jsModule)
		if err != nil {
			return err
		}
	} else {
		return errors.New("invalid JS module")
	}

	return nil
}
