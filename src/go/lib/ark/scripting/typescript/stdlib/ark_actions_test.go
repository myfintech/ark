package stdlib

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/ark/storage/memory"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems/http_server"
	"github.com/myfintech/ark/src/go/lib/logz"
	"github.com/stretchr/testify/require"
)

func TestActionsBuildDockerImage(t *testing.T) {
	store := new(memory.Store)
	logger := new(logz.NoOpLogger)
	broker := new(cqrs.NoOpBroker)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	man := subsystems.NewManager(ctx)

	err := man.Register(http_server.NewSubsystem("127.0.0.1:9000", "", store, logger, broker))
	require.NoError(t, err)

	err = man.Start()
	require.NoError(t, err)

	require.NoError(t, err)

	module, err := vm.ResolveModule(filepath.Join(testdata, "src/service/service-foo/build.ts"))
	require.NoError(t, err)

	jsonBytes, err := module.ToObject(vm.Runtime).MarshalJSON()
	require.NoError(t, err)
	t.Log(string(jsonBytes))
}
