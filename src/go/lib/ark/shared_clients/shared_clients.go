package shared_clients

import (
	vault "github.com/hashicorp/vault/api"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/protocols/watermill/gochannel"
	"github.com/myfintech/ark/src/go/lib/ark/kv"
	"github.com/myfintech/ark/src/go/lib/ark/workspace"
	"github.com/myfintech/ark/src/go/lib/container"
	"github.com/myfintech/ark/src/go/lib/kube"
	"github.com/myfintech/ark/src/go/lib/logz"
	"github.com/myfintech/ark/src/go/lib/vault_tools"
)

// K8sClientUser an interface that injects kube.Client
type K8sClientUser interface {
	UseK8sClient(client kube.Client)
}

// DockerClientUser an interface that injects container.Docker
type DockerClientUser interface {
	UseDockerClient(client container.Docker)
}

// VaultClientUser an interface that injects a vault.Client
type VaultClientUser interface {
	UseVaultClient(client *vault.Client)
}

// KVStorageUser an interface that injects kv.Storage
type KVStorageUser interface {
	UseKVStorage(client kv.Storage)
}

// WorkspaceConfigUser an interface that injects a blob storage URL
type WorkspaceConfigUser interface {
	UseWorkspaceConfig(config workspace.Config)
}

// BrokerUser an interface that injects kv.Storage
type BrokerUser interface {
	UseBroker(client cqrs.Broker)
}

// Container a shared client dependency injection container
// FIXME(erick): replace pointers with interface
type Container struct {
	K8s             kube.Client
	Vault           *vault.Client
	Docker          container.Docker
	KVStorage       kv.Storage
	Broker          cqrs.Broker `wire:"-"`
	Logger          logz.FieldLogger
	WorkspaceConfig workspace.Config
}

// Inject accepts any interface and checks if it it implements one of the client user interfaces in this library
// The error can be safely ignored unless you want to verify that the value was an injectable interface
func (c *Container) Inject(v interface{}) {
	if t, ok := v.(K8sClientUser); ok {
		t.UseK8sClient(c.K8s)
	}

	if t, ok := v.(DockerClientUser); ok {
		t.UseDockerClient(c.Docker)
	}

	if t, ok := v.(VaultClientUser); ok {
		t.UseVaultClient(c.Vault)
	}

	if t, ok := v.(KVStorageUser); ok {
		t.UseKVStorage(c.KVStorage)
	}

	if t, ok := v.(WorkspaceConfigUser); ok {
		t.UseWorkspaceConfig(c.WorkspaceConfig)
	}

	if t, ok := v.(BrokerUser); ok {
		t.UseBroker(c.Broker)
	}

	if t, ok := v.(logz.Injector); ok {
		t.UseLogger(c.Logger)
	}
}

func NewContainerWithDefaults() (*Container, error) {
	vaultClient, err := vault_tools.InitClient(nil)
	if err != nil {
		return nil, err
	}

	dockerClient, err := container.NewDockerClient(container.DefaultDockerCLIOptions()...)
	if err != nil {
		return nil, err
	}

	return &Container{
		K8s:    kube.Init(nil),
		Vault:  vaultClient,
		Docker: *dockerClient,
		Broker: gochannel.New(),
		KVStorage: &kv.VaultStorage{
			Client:        vaultClient,
			FSBasePath:    "",
			EncryptionKey: "",
		},
		Logger: logz.NoOpLogger{},
	}, nil
}
