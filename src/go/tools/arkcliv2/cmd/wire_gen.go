// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package cmd

import (
	"github.com/moby/buildkit/util/appcontext"
	"github.com/myfintech/ark/src/go/lib/ark/shared_clients"
	"github.com/myfintech/ark/src/go/lib/ark/storage/memory"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems"
	"github.com/myfintech/ark/src/go/lib/ark/workspace"

	_ "embed"
)

// Injectors from wire.go:

func BuildCLI() (*ArkCLI, error) {
	command := newRootCmd()
	writer, err := newLogger(command)
	if err != nil {
		return nil, err
	}
	config, err := workspace.LoadConfigFromCWD()
	if err != nil {
		return nil, err
	}
	client, err := newK8sClient(config)
	if err != nil {
		return nil, err
	}
	apiClient, err := newVaultClient(config)
	if err != nil {
		return nil, err
	}
	docker, err := newDockerClient()
	if err != nil {
		return nil, err
	}
	storage := newVaultKVStore(apiClient, config)
	workspaceConfig := newWorkspaceCopy(config)
	container := &shared_clients.Container{
		K8s:             client,
		Vault:           apiClient,
		Docker:          docker,
		KVStorage:       storage,
		Logger:          writer,
		WorkspaceConfig: workspaceConfig,
	}
	http_serverClient := newServerClient()
	v, err := newGitIgnorePatterns(config)
	if err != nil {
		return nil, err
	}
	context := appcontext.Context()
	watchmanClient, err := newWatchmanClient(context)
	if err != nil {
		return nil, err
	}
	virtualMachine, err := newTypeScriptVM(config, http_serverClient, storage, v, watchmanClient)
	if err != nil {
		return nil, err
	}
	manager := subsystems.NewManager(context)
	proc, err := newHostServerDaemon()
	if err != nil {
		return nil, err
	}
	observer, err := newFSObserver(writer, config, v, watchmanClient)
	if err != nil {
		return nil, err
	}
	store := memory.NewCachedMemoryStore()
	cmdCoreConfig := coreConfig{
		logger:            writer,
		config:            config,
		sharedClients:     container,
		vm:                virtualMachine,
		vaultClient:       apiClient,
		subsystemsManager: manager,
		hostServerDaemon:  proc,
		fileObserver:      observer,
		store:             store,
		kvStorage:         storage,
		httpClient:        http_serverClient,
		gitIgnorePatterns: v,
	}
	arkCLI, err := newArkCLI(command, cmdCoreConfig)
	if err != nil {
		return nil, err
	}
	return arkCLI, nil
}
