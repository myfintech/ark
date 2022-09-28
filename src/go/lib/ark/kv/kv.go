package kv

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"k8s.io/kubectl/pkg/util/term"

	gonanoid "github.com/matoous/go-nanoid"

	"github.com/myfintech/ark/src/go/lib/utils"

	vault "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

// Storage is the interface that implements the Get and Put methods of the VaultStorage struct
type Storage interface {
	Get(path string) (secretData map[string]interface{}, err error)
	Put(path string, mergeValues map[string]interface{}) (map[string]interface{}, error)
	Edit(path string) error
	EncryptedDataPath() (kvDataPath string)
	DecryptedDataPath() (tmpPath string)
	DecryptToFile(secretPath string) (secretFile string, err error)
}

// VaultStorage holds the configuration using Vault transit via the kv library
type VaultStorage struct {
	Client        *vault.Client
	FSBasePath    string
	EncryptionKey string
}

// The Get method reads configuration from the local Vault path
func (s *VaultStorage) Get(path string) (map[string]interface{}, error) {
	config := make(map[string]interface{})

	vaultFile := filepath.Join(s.FSBasePath, path)

	if _, err := os.Stat(vaultFile); err != nil {
		return config, err
	}
	vaultFileBytes, bytesErr := ioutil.ReadFile(vaultFile)
	if bytesErr != nil {
		return config, errors.Wrapf(bytesErr, "failed to read %s", vaultFile)
	}

	vaultSecret, secretErr := s.Client.Logical().Write(fmt.Sprintf("transit/decrypt/%s", s.EncryptionKey), map[string]interface{}{
		"ciphertext": string(vaultFileBytes),
	})
	if secretErr != nil {
		return config, secretErr
	}

	base64EncodedVaultSecret := vaultSecret.Data["plaintext"].(string)

	vaultSecretDecodedBase64, decodeErr := base64.StdEncoding.DecodeString(base64EncodedVaultSecret)
	if decodeErr != nil {
		return config, decodeErr
	}

	if err := json.Unmarshal(vaultSecretDecodedBase64, &config); err != nil {
		return config, errors.Wrapf(err, "failed to unmarshal secret to config: %v", &config)
	}

	return config, nil
}

// The Put method writes configuration to the local Vault path
func (s *VaultStorage) Put(path string, mergeValues map[string]interface{}) (map[string]interface{}, error) {
	config, err := s.Get(path)
	if err != nil && !os.IsNotExist(err) {
		return config, err
	}

	vaultFile := filepath.Join(s.FSBasePath, path)

	config = utils.MergeMaps(config, mergeValues)

	hydratedConfig, err := json.Marshal(config)
	if err != nil {
		return config, errors.Wrapf(err, "failed to marshal config: %v", &config)
	}

	encodedConfig := base64.StdEncoding.EncodeToString(hydratedConfig)

	encryptedConfig, encryptErr := s.Client.Logical().Write(fmt.Sprintf("transit/encrypt/%s", s.EncryptionKey), map[string]interface{}{
		"plaintext": encodedConfig,
	})
	if encryptErr != nil {
		return config, encryptErr
	}

	cipherText := encryptedConfig.Data["ciphertext"].(string)

	if err = os.MkdirAll(filepath.Dir(vaultFile), 0755); err != nil {
		return config, errors.Wrapf(err, "failed to make directory path")
	}

	if err = ioutil.WriteFile(vaultFile, []byte(cipherText), 0644); err != nil {
		return config, errors.Wrapf(err, "failed to write file to: %s", vaultFile)
	}

	return config, nil
}

// Edit opens a workspace KV file in the user's $EDITOR for modifications and applies those modifications
func (s *VaultStorage) Edit(secretPath string) error {
	editor := utils.EnvLookup("EDITOR", "vi")
	editorWithFlags := strings.Split(editor, " ")

	tmpPath, err := s.DecryptToFile(secretPath)
	if err != nil {
		return err
	}

	cmd := exec.Command(editorWithFlags[0], append(editorWithFlags[1:], tmpPath)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err = (term.TTY{In: os.Stdin, TryDev: true}).Safe(cmd.Run); err != nil {
		return err
	}

	newContentForPut := make(map[string]interface{})
	fileBytes, err := ioutil.ReadFile(tmpPath)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(fileBytes, &newContentForPut); err != nil {
		return err
	}

	if _, err = s.Put(secretPath, newContentForPut); err != nil {
		return err
	}

	return nil
}

// EncryptedDataPath returns the base path of VaultStorage
func (s *VaultStorage) EncryptedDataPath() string {
	return s.FSBasePath
}

// DecryptedDataPath method return the temp path to decrypt secrets to
func (s *VaultStorage) DecryptedDataPath() string {
	return os.TempDir()
}

// DecryptToFile writes a secret to a random secret data file on disk
func (s VaultStorage) DecryptToFile(secret string) (path string, err error) {
	// generate random secret file name
	randName, err := gonanoid.Nanoid(32)
	if err != nil {
		return
	}

	randName += ".json"
	path = filepath.Join(s.DecryptedDataPath(), randName)

	// create random secret file under a temp file system
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return
	}

	defer func() {
		_ = file.Close()
	}()

	// get the secret from encrypted storage
	data, err := s.Get(secret)
	if err != nil {
		return
	}

	// marshal the secret into JSON
	jsonBytes, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return
	}

	// create a reader from the in memory JSON
	reader := bytes.NewReader(jsonBytes)

	// copy the secret data to the random temp file
	// if it fails close the file and remove its path
	if _, err = io.Copy(file, reader); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return
	}

	return
}
