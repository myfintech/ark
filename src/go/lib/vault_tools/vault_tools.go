package vault_tools

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/myfintech/ark/src/go/lib/utils"

	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/myfintech/ark/src/go/lib/log"
)

// SecretDataPath format a set of paths into vaults secret/data path
func SecretDataPath(path ...string) string {
	return filepath.Join(append([]string{"secret", "data"}, path...)...)
}

// PrefixDataPath ...

// SecretPath format a set of paths into vaults secret/ path

// SecretMetaPath format a set of paths into vaults secret/metadata path
func SecretMetaPath(path ...string) string {
	return filepath.Join(append([]string{"secret", "metadata"}, path...)...)
}

// ExtractListData builds a set of paths from a directory in vault
func ExtractListData(secret *api.Secret) ([]string, bool) {
	if secret == nil || secret.Data == nil {
		return nil, false
	}
	k, ok := secret.Data["keys"]
	if !ok || k == nil {
		return nil, false
	}
	paths, ok := k.([]interface{})
	if !ok || paths == nil {
		return nil, false
	}
	list := make([]string, 0, len(paths))
	for _, p := range paths {
		if str, listOk := p.(string); listOk {
			list = append(list, str)
		}
	}
	return list, ok
}

// FindAllSecretsRecursive crawls vault at a given starting location and returns a set of full vault paths of every secret found
func FindAllSecretsRecursive(client *api.Client, path []string) (secretPaths []string, err error) {
	resp, err := client.Logical().List(SecretMetaPath(path...))
	if err != nil {
		return
	}
	if resp == nil {
		return nil, errors.Errorf("there was no secret at %s", SecretMetaPath(path...))
	}

	keys, ok := ExtractListData(resp)
	if !ok {
		return nil, errors.Errorf("keys were not of expected type []string at %s", SecretMetaPath(path...))
	}

	var next []string

	for _, k := range keys {
		next = append(path, k)
		if strings.HasSuffix(k, "/") {
			subPaths, subErr := FindAllSecretsRecursive(client, next)
			if subErr != nil {
				return secretPaths, subErr
			}
			secretPaths = append(secretPaths, subPaths...)
		} else {
			secretPaths = append(secretPaths, SecretDataPath(next...))
		}
	}
	return
}

type FieldMapping struct {
	Src  string `json:"src"`
	Dst  string `json:"dst"`
	Drop bool   `json:"drop"`
}

type CopySecretOptions struct {
	queue     chan string
	Src       string
	Dst       string
	Mappings  []FieldMapping
	Overwrite bool
	Client    *api.Client
	Threads   int
}

func CopySecret(opts CopySecretOptions) func() error {
	return func() error {
		srcPrefix := SecretDataPath(opts.Src)
		dstPrefix := SecretDataPath(opts.Dst)

		for secretPath := range opts.queue {
			secret, err := opts.Client.Logical().Read(secretPath)
			if err != nil {
				return err
			}

			if secret == nil {
				return errors.Errorf("src doesn't exist %s", secretPath)
			}

			dstSecretPath := strings.Replace(secretPath, srcPrefix, dstPrefix, 1)

			dstSecret, err := opts.Client.Logical().Read(dstSecretPath)
			if err != nil {
				return err
			}
			if dstSecret != nil && !opts.Overwrite {
				log.Warnf("dst already exists at %s, skipping", dstSecretPath)
				continue
			}

			if opts.Mappings != nil {
				data := secret.Data["data"].(map[string]interface{})
				for _, mapping := range opts.Mappings {
					srcVal, exists := data[mapping.Src]
					if !exists {
						return errors.Errorf("failed to remap field %s, no match found", mapping.Src)
					}
					data[mapping.Dst] = srcVal
					if mapping.Drop {
						delete(data, mapping.Src)
					}
				}
				secret.Data["data"] = data
			}

			log.Infof("copying from %s to %s", secretPath, dstSecretPath)
			if _, err = opts.Client.Logical().Write(dstSecretPath, secret.Data); err != nil {
				return err
			}
		}
		return nil
	}
}

func InitClient(c *api.Config) (client *api.Client, err error) {
	client, err = api.NewClient(c)
	if err != nil {
		return
	}

	err = client.SetAddress(os.Getenv(api.EnvVaultAddress))
	if err != nil {
		return
	}

	if client.Token() != "" {
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	tokenEnv := os.Getenv("VAULT_TOKEN")
	if tokenEnv != "" {
		client.SetToken(tokenEnv)
		return
	}
	data, err := ioutil.ReadFile(filepath.Join(home, ".vault-token"))
	if err != nil {
		log.Warn(err)
	} else {
		client.SetToken(string(data))
		return
	}
	configData, err := os.ReadFile(utils.EnvLookup("VAULT_CONFIG", "/etc/vault/config.json"))
	if err != nil {
		return
	}
	var token map[string]string
	err = json.Unmarshal(configData, &token)
	if err != nil {
		return
	}
	client.SetToken(token["token"])
	return

}

const (
	vaultDefaultConfigVar = "VAULT_DEFAULT_CONFIG"
	vaultEnvVar           = "VAULT_ENV"
	vaultTeamVar          = "VAULT_TEAM"
)

// VaultConfig defines the structure of a Vault token stored in /etc/vault/config.json
type VaultConfig struct {
	Token string `json:"token"`
}
