package stdlib

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/myfintech/ark/src/go/lib/watchman"

	"github.com/myfintech/ark/src/go/lib/ark/kv"
	"github.com/myfintech/ark/src/go/lib/vault_tools/vault_test_harness"

	"github.com/myfintech/ark/src/go/lib/ark/subsystems/http_server"
	"github.com/myfintech/ark/src/go/lib/git/gitignore"

	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript"
)

var cwd string
var testdata string
var vm *typescript.VirtualMachine
var storage kv.Storage

func TestMain(m *testing.M) {
	var err error
	vm = typescript.MustInitVM(nil)

	cwd, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	testdata = filepath.Join(cwd, "testdata")
	client := http_server.NewClient("http://127.0.0.1:9000")
	t := &testing.T{}

	vaultClient, cleanup := vault_test_harness.CreateVaultTestCore(t, false)
	defer cleanup()

	storage = &kv.VaultStorage{
		Client:        vaultClient,
		FSBasePath:    filepath.Join(testdata, "src/kv/.ark/kv"),
		EncryptionKey: "mantl-key",
	}

	wm, err := watchman.Connect(context.Background(), 10)
	if err != nil {
		panic(err)
	}

	defer func(wm *watchman.Client) {
		_, err := wm.DeleteAll()
		if err != nil {
			require.NoError(t, err)
		}
	}(wm)

	_, err = wm.WatchProject(watchman.WatchProjectOptions{Directory: testdata})
	if err != nil {
		panic(err)
	}

	err = vm.InstallModuleListWithPrefix("arksdk", typescript.ModuleList{
		"actions": NewArkActionLibrary(Options{
			FSRealm:   testdata,
			Runtime:   vm.Runtime,
			Client:    client,
			GitIgnore: gitignore.NewMatcher(nil),
		}),
		"filepath": NewFilepathLibrary(Options{
			FSRealm:        testdata,
			Runtime:        vm.Runtime,
			Client:         client,
			GitIgnore:      gitignore.NewMatcher(nil),
			WatchmanClient: wm,
		}),
		"kv": NewKvLibrary(KvLibraryOptions{
			Runtime:   vm.Runtime,
			KVStorage: storage,
		}),
		"encoding": NewEncodingLibrary(EncodingLibraryOptions{
			Runtime: vm.Runtime,
		}),
	})
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}
