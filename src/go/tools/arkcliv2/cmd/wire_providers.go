package cmd

import (
	"context"
	"os"
	"path/filepath"

	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems"
	"github.com/myfintech/ark/src/go/lib/watchman"

	"github.com/myfintech/ark/src/go/lib/ark/scripting/typescript/container_plugins"
	"github.com/myfintech/ark/src/go/lib/ark/scripting/typescript/plugins"
	"github.com/myfintech/ark/src/go/lib/ark/scripting/typescript/stdlib"

	gitignorev5 "github.com/go-git/go-git/v5/plumbing/format/gitignore"

	"github.com/myfintech/ark/src/go/lib/git/gitignore"

	"github.com/myfintech/ark/src/go/lib/container"

	"github.com/myfintech/ark/src/go/lib/vault_tools"

	"github.com/myfintech/ark/src/go/lib/kube"

	"github.com/myfintech/ark/src/go/lib/ark/storage/memory"

	"github.com/spf13/cobra"

	"github.com/google/wire"
	vault "github.com/hashicorp/vault/api"
	"github.com/moby/buildkit/util/appcontext"
	"github.com/myfintech/ark/src/go/lib/ark/kv"
	"github.com/myfintech/ark/src/go/lib/ark/shared_clients"
	"github.com/myfintech/ark/src/go/lib/ark/subsystems/http_server"
	"github.com/myfintech/ark/src/go/lib/ark/workspace"
	"github.com/myfintech/ark/src/go/lib/daemonize"
	"github.com/myfintech/ark/src/go/lib/embedded_scripting/typescript"
	"github.com/myfintech/ark/src/go/lib/fs/observer"
	"github.com/myfintech/ark/src/go/lib/logz"
	"github.com/myfintech/ark/src/go/lib/logz/transports"
)

// creates concrete alias for wire to identify this dependency
type workingDir string

// a container for all dependencies
// TODO: break this down, this is phase 1 of moving away globals
type coreConfig struct {
	logger            *logz.Writer
	config            *workspace.Config
	sharedClients     *shared_clients.Container
	vm                *typescript.VirtualMachine
	vaultClient       *vault.Client
	subsystemsManager *subsystems.Manager
	hostServerDaemon  *daemonize.Proc
	fileObserver      *observer.Observer
	store             ark.Store
	kvStorage         kv.Storage
	httpClient        http_server.Client
	gitIgnorePatterns []gitignorev5.Pattern
}

func newCWD(cmd *cobra.Command) (workingDir, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return workingDir(cwd), err
	}

	cwd, err = cmd.Flags().GetString("cwd")
	if err != nil {
		return workingDir(cwd), err
	}
	return workingDir(cwd), err
}

func newLogger(cmd *cobra.Command) (*logz.Writer, error) {
	levelStr, err := cmd.PersistentFlags().GetString("log-level")
	if err != nil {
		return nil, err
	}

	logger := logz.New(
		appcontext.Context(),
		logz.WithLevelString(levelStr),
		logz.WithMux(transports.DefaultIOWriter),
		logz.WithMux(transports.SuggestedLogFileWriter(
			"ark/",
			"server.log"),
		),
	)

	if err = logger.InitError(); err != nil {
		return nil, err
	}

	return logger, nil
}

func newVaultClient(config *workspace.Config) (*vault.Client, error) {
	vaultClient, err := vault_tools.InitClient(nil)
	if err != nil {
		return nil, err
	}

	if vaultClient.Address() == "" && config.Vault.Address != "" {
		if err = vaultClient.SetAddress(config.Vault.Address); err != nil {
			return nil, err
		}
	}
	return vaultClient, nil
}

func newGitIgnorePatterns(config *workspace.Config) ([]gitignorev5.Pattern, error) {
	patterns, err := gitignore.LoadRepoPatterns(config.Root())
	if err != nil {
		return patterns, err
	}
	return patterns, nil
}

func newServerClient() http_server.Client {
	return http_server.NewClient("127.0.0.1:9000")
}

func newVaultKVStore(vaultClient *vault.Client, config *workspace.Config) kv.Storage {
	return &kv.VaultStorage{
		Client:        vaultClient,
		FSBasePath:    filepath.Join(config.Dir(), "kv"),
		EncryptionKey: config.Vault.EncryptionKey,
	}
}

func newDockerClient() (container.Docker, error) {
	d, err := container.NewDockerClient(container.DefaultDockerCLIOptions()...)
	if err != nil {
		return container.Docker{}, err
	}
	return *d, nil
}

func newK8sClient(config *workspace.Config) (kube.Client, error) {
	client, err := kube.InitWithSafeContexts("", config.K8s.GetSafeContexts())
	if err != nil {
		return client, err
	}
	return client, nil
}

func newWatchmanClient(ctx context.Context) (*watchman.Client, error) {
	wm, err := watchman.Connect(ctx, 10)
	if err != nil {
		return &watchman.Client{}, nil
	}
	return wm, nil
}

func newTypeScriptVM(
	config *workspace.Config,
	serverClient http_server.Client,
	kvStorage kv.Storage,
	gitIgnorePatterns []gitignorev5.Pattern,
	watchmanClient *watchman.Client,
) (*typescript.VirtualMachine, error) {
	// init js vm
	vm := typescript.MustInitVM([]typescript.Library{
		{
			Prefix: "ark/external",
			Path:   filepath.Join(config.Dir(), "external_modules"),
		},
		{
			Prefix: "ark/native",
			Path:   filepath.Join(config.Dir(), "native_modules"),
		},
	})

	err := plugins.Load(vm, plugins.NewLibrary(stdlib.Options{
		Runtime: vm.Runtime,
	}))
	if err != nil {
		return vm, err
	}

	err = vm.InstallModuleListWithPrefix("arksdk", typescript.ModuleList{
		"actions": stdlib.NewArkActionLibrary(stdlib.Options{
			FSRealm:        config.Root(),
			Runtime:        vm.Runtime,
			Client:         serverClient,
			GitIgnore:      gitignore.NewMatcher(gitIgnorePatterns),
			WatchmanClient: watchmanClient,
		}),
		"filepath": stdlib.NewFilepathLibrary(stdlib.Options{
			FSRealm:        config.Root(),
			Runtime:        vm.Runtime,
			Client:         serverClient,
			GitIgnore:      gitignore.NewMatcher(gitIgnorePatterns),
			WatchmanClient: watchmanClient,
		}),
		"kv": stdlib.NewKvLibrary(stdlib.KvLibraryOptions{
			KVStorage: kvStorage,
			Runtime:   vm.Runtime,
		}),
		"encoding": stdlib.NewEncodingLibrary(stdlib.EncodingLibraryOptions{
			Runtime: vm.Runtime,
		}),
	})
	if err != nil {
		return nil, err
	}

	if err = container_plugins.Load(appcontext.Context(), vm, config.Plugins); err != nil {
		return nil, err
	}
	return vm, nil
}

func newHostServerDaemon() (*daemonize.Proc, error) {
	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}

	return daemonize.NewProc(
		executable,
		[]string{"server", "run"},
		filepath.Join(os.TempDir(), "ark", "server", "pid"),
	), nil
}

func newFSObserver(
	logger logz.FieldLogger,
	config *workspace.Config,
	gitIgnorePatterns []gitignorev5.Pattern,
	watchmanClient *watchman.Client,
) (*observer.Observer, error) {
	socket, _ := watchman.GetSocketName()
	nativeModeEnabled := socket == ""
	if nativeModeEnabled {
		logger.Infof("fs.rx.observer running in native mode")
	} else {
		logger.Infof("fs.rx.observer running in watchman mode")
	}

	logger = logger.Child(logz.WithFields(logz.Fields{
		"system": "fs.rx.observer",
	}))

	return observer.NewObserverWithoutIndexing(
		nativeModeEnabled,
		true,
		config.Root(),
		config.FileSystem.Ignore,
		gitignore.NewMatcher(gitIgnorePatterns),
		logger,
		watchmanClient,
	), nil
}

func newWorkspaceCopy(config *workspace.Config) workspace.Config {
	return *config
}

var baseSet = wire.NewSet(
	newCWD,
	newRootCmd,
	newLogger,
	newGitIgnorePatterns,
	newTypeScriptVM,
	newHostServerDaemon,
	newFSObserver,
	appcontext.Context,
	subsystems.NewManager,
	workspace.LoadConfigFromCWD,
	memory.NewCachedMemoryStore,
	newWorkspaceCopy,
	wire.Bind(new(ark.Store), new(*memory.Store)),
	wire.Bind(new(logz.FieldLogger), new(*logz.Writer)),
)

var clientSet = wire.NewSet(
	newVaultClient,
	newVaultKVStore,
	newServerClient,
	newK8sClient,
	newDockerClient,
	newWatchmanClient,
	wire.Struct(new(shared_clients.Container), "*"),
)

var coreSet = wire.NewSet(
	baseSet,
	clientSet,
	newArkCLI,
	wire.Struct(new(coreConfig), "*"),
)
