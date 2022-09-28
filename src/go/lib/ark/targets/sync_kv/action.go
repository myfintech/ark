package sync_kv

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/myfintech/ark/src/go/lib/vault_tools"
	"golang.org/x/sync/errgroup"

	"github.com/myfintech/ark/src/go/lib/ark/kv"
	"github.com/myfintech/ark/src/go/lib/fs"

	"github.com/pkg/errors"

	vault "github.com/hashicorp/vault/api"
	intlnet "github.com/myfintech/ark/src/go/lib/internal_net"
	"github.com/myfintech/ark/src/go/lib/log"
)

// Action is the executor for implementing a KV Sync
type Action struct {
	Artifact           *Artifact
	Target             *Target
	KVStorage          kv.Storage
	VaultClientFactory VaultClientFactory
	Logger             logz.FieldLogger
}

var _ logz.Injector = &Action{}

// UseLogger injects a logger into the target's action
func (a *Action) UseLogger(logger logz.FieldLogger) {
	a.Logger = logger
}

// UseKVStorage inject the client
func (a *Action) UseKVStorage(client kv.Storage) {
	a.KVStorage = client
}

// VaultClientFactory allows for the injection of a custom configured Vault client
type VaultClientFactory func(config *vault.Config) (*vault.Client, error)

// Execute runs the action and produces a sync_kv.Artifact
func (a Action) Execute(ctx context.Context) (err error) {
	maxRetries := 5
	timeout := 60 * time.Second

	if a.Target.TimeoutSeconds != 0 {
		timeout = time.Duration(a.Target.TimeoutSeconds) * time.Second
	}

	if a.Target.MaxRetries > 0 {
		maxRetries = a.Target.MaxRetries
	}

	var clientFactory VaultClientFactory

	if a.VaultClientFactory != nil {
		clientFactory = a.VaultClientFactory
	} else {
		clientFactory = vault.NewClient
	}

	client, err := clientFactory(&vault.Config{
		Address: a.Target.EngineURL,
		Timeout: timeout,
	})
	if err != nil {
		return err
	}

	client.SetToken(a.Target.Token)

	addr, err := url.Parse(client.Address())
	if err != nil {
		return err
	}

	probeOptions := intlnet.ProbeOptions{
		Timeout:        timeout,
		Delay:          time.Second * 3,
		ExpectedStatus: 0,
		Address:        addr,
		MaxRetries:     maxRetries,
		OnError: func(err error, remainingAttempts int) {
			log.Warnf("waiting for Vault to initialize and unseal: %v, remaining attempts %d",
				err, remainingAttempts)
		},
	}

	// https://vaultproject.io/api/system/health
	// 200 == initialized, unsealed, and active (normally)
	if err = intlnet.RunProbe(probeOptions, func() error {
		health, hErr := client.Sys().Health()
		if hErr != nil {
			return hErr
		}
		if health.Initialized && !health.Sealed {
			return nil
		}
		return errors.New("not healthy")
	}); err != nil {
		return err
	}

	secretPrefix := a.KVStorage.EncryptedDataPath()

	queue := make(chan string, 100)
	eg, egctx := errgroup.WithContext(ctx)

	for i := 0; i < 50; i++ {
		eg.Go(copyToVault(a.KVStorage, client, queue))
	}

	eg.Go(func() error {
		defer close(queue)
		for _, file := range a.Target.SourceFiles {
			if !strings.HasPrefix(file, secretPrefix) {
				return errors.Errorf("the source file %s does not match the prefix: %s", file, secretPrefix)
			}

			select {
			case queue <- fs.TrimPrefix(file, secretPrefix):
			case <-egctx.Done():
				return nil
			}
		}
		return nil
	},
	)

	return eg.Wait()
}

func copyToVault(storage kv.Storage, client *vault.Client, queue chan string) func() error {
	return func() error {
		for trimmedPath := range queue {
			data, err := storage.Get(trimmedPath)
			if err != nil {
				return err
			}

			_, err = client.Logical().Write(vault_tools.SecretDataPath(trimmedPath), map[string]interface{}{
				"data": data,
			})
			if err != nil {
				return err
			}
		}
		return nil
	}
}
