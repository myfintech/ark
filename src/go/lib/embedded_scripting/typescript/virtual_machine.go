package typescript

import (
	"bytes"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"

	"github.com/dop251/goja"
)

// Module is an alias type that can be installed in the module resolver
type Module map[string]interface{}

// ModuleList a structure used to install many modules into the typescript runtime quickly
type ModuleList map[string]Module

// VirtualMachine an opinionated wrapper around the goja Runtime that provides native support for typescript
type VirtualMachine struct {
	transpiler     *Transpiler
	moduleResolver *ModuleResolver
	Runtime        *goja.Runtime
	mutex          sync.Mutex
}

// InstallPlugins allows the typescript Runtime to be extended with plugins
func (vm *VirtualMachine) InstallPlugins(plugins []Plugin) error {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()
	return InstallPlugins(vm.Runtime, plugins)
}

// ResolveModule resolve, transpiles, and caches the given typescript file
func (vm *VirtualMachine) ResolveModule(filename string) (*goja.Object, error) {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()
	return vm.moduleResolver.resolveAndTranspile(filename)
}

// GetModule ...
func (vm *VirtualMachine) GetModule(name string) (*goja.Object, error) {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()
	return vm.moduleResolver.get(name), nil
}

// RunScript executes the given src in the global context.
// uses the filename for source mapping
func (vm *VirtualMachine) RunScript(filename, src string) (goja.Value, error) {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()
	ts, err := vm.transpiler.Transpile(bytes.NewBufferString(src))
	if err != nil {
		return nil, err
	}
	return vm.Runtime.RunScript(filename, ts)
}

// CreateNewRuntimeObject creates a new *goja.Object using the VMs Runtime
func (vm *VirtualMachine) CreateNewRuntimeObject() *goja.Object {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()
	return vm.Runtime.NewObject()
}

// SetRuntimeValue sets a variable by name and value in the VMs Runtime
func (vm *VirtualMachine) SetRuntimeValue(name string, value interface{}) error {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()
	return vm.Runtime.Set(name, value)
}

// InstallModule installs a native module in the resolver allows creating libraries in go that can be imported or required
// returns an error if a module with the same name already exits in the VM's ModuleResolver cache
func (vm *VirtualMachine) InstallModule(name string, module Module) error {
	if mod := vm.moduleResolver.get(name); mod != nil {
		return errors.Errorf("a module with name %s already exists", name)
	}

	mod := vm.Runtime.NewObject()
	if err := mod.Set("exports", module); err != nil {
		return errors.Wrapf(err, "failed to install module %s", name)
	}

	vm.moduleResolver.set(name, mod)
	return nil
}

// InstallModuleListWithPrefix uses InstallModule to quickly install a large number of modules and expose them in different ways
// Example:
// 		import * as arksdk from 'arksdk'
//		import * as filepath from 'arksdk/filepath'
func (vm *VirtualMachine) InstallModuleListWithPrefix(prefix string, modules ModuleList) error {
	module := make(Module)
	for name, mod := range modules {
		module[name] = mod
		fullname := filepath.Join(prefix, name)
		if err := vm.InstallModule(fullname, mod); err != nil {
			return errors.Wrapf(err, "failed to install %s", fullname)
		}
	}

	if err := vm.InstallModule(prefix, module); err != nil {
		return errors.Wrapf(err, "failed to install root module %s", prefix)
	}
	return nil
}

// NewVirtualMachine creates a new typescript virtual machine
func NewVirtualMachine(libraries []Library) (vm *VirtualMachine, err error) {
	runtime := goja.New()
	runtime.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	transpiler, err := NewTranspiler()
	if err != nil {
		return
	}

	resolver := &ModuleResolver{
		Transpiler:      transpiler,
		CompilerOptions: DefaultCompilerOptions,
		Libraries:       libraries,
	}

	err = InstallPlugins(runtime, []Plugin{
		transpiler,
		resolver,
	})
	if err != nil {
		return
	}

	return &VirtualMachine{
		transpiler:     transpiler,
		moduleResolver: resolver,
		Runtime:        runtime,
	}, nil
}

// MustInitVM initialize a JS or panic
func MustInitVM(libraries []Library) *VirtualMachine {
	vm, err := NewVirtualMachine(libraries)
	if err != nil {
		panic(err)
	}

	return vm
}
