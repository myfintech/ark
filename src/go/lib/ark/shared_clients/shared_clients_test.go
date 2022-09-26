package shared_clients

import (
	"testing"

	"github.com/myfintech/ark/src/go/lib/ark/workspace"

	vault "github.com/hashicorp/vault/api"

	"github.com/myfintech/ark/src/go/lib/ark/kv"

	"github.com/myfintech/ark/src/go/lib/container"

	"github.com/myfintech/ark/src/go/lib/kube"
	"github.com/stretchr/testify/mock"
)

type mockClientUser struct {
	mock.Mock
}

func (m *mockClientUser) UseVaultClient(client *vault.Client) {
	m.Called(client)
}

func (m *mockClientUser) UseKVStorage(client kv.Storage) {
	m.Called(client)
}

func (m *mockClientUser) UseDockerClient(client container.Docker) {
	m.Called(client)
}

func (m *mockClientUser) UseK8sClient(client kube.Client) {
	m.Called(client)
}

func (m *mockClientUser) UseWorkspaceConfig(config workspace.Config) {
	m.Called(config)
}

func TestContainer(t *testing.T) {
	user := new(mockClientUser)
	sharedLibsContainer := new(Container)

	user.On("UseK8sClient", mock.Anything)
	user.On("UseDockerClient", mock.Anything)
	user.On("UseVaultClient", mock.Anything)
	user.On("UseKVStorage", mock.Anything)
	user.On("UseWorkspaceConfig", mock.Anything)

	sharedLibsContainer.Inject(user)

	user.AssertExpectations(t)
}
