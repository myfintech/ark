package plugins

import (
	"path/filepath"

	"github.com/myfintech/ark/src/go/lib/ark/scripting/typescript/stdlib"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript"
	"github.com/pkg/errors"
)

type NativePlugin struct {
	Name   string
	Module typescript.Module
}

func NewLibrary(opts stdlib.Options) []NativePlugin {
	return []NativePlugin{
		NewKafkaPlugin(opts),
		NewVaultPlugin(opts),
		NewPostgresPlugin(opts),
		NewMicroservicePlugin(opts),
		NewRedisPlugin(opts),
		NewGcloudEmulatorPlugin(opts),
		NewCiCdGhaRunnerPlugin(opts),
		NewConsulPlugin(opts),
		NewCoreProxyPlugin(opts),
		NewDatadogPlugin(opts),
		NewKubeStatePlugin(opts),
		NewNsReaperPlugin(opts),
		NewSDMPlugin(opts),
		NewTerraformCloudAgentPlugin(opts),
		NewVaultServiceAccountPlugin(opts),
		NewVDSPlugin(opts),
	}
}

func Load(vm *typescript.VirtualMachine, plugins []NativePlugin) error {
	for _, plugin := range plugins {
		if err := vm.InstallModule(filepath.Join("ark/plugins", plugin.Name), plugin.Module); err != nil {
			return errors.Errorf("got an error installing module %s, %v", plugin.Name, err)
		}
	}

	return nil
}
