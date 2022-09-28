package typescript

import (
	"github.com/dop251/goja"
)

type Plugin interface {
	Install(vm *goja.Runtime) error
}

func InstallPlugins(runtime *goja.Runtime, plugins []Plugin) (err error) {
	for _, plugin := range plugins {
		if err = plugin.Install(runtime); err != nil {
			return
		}
	}
	return
}
