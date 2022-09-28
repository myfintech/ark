package docker_image

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/myfintech/ark/src/go/lib/ark/kv"

	"github.com/myfintech/ark/src/go/lib/vault_tools/vault_test_harness"

	"github.com/myfintech/ark/src/go/lib/ark"

	"github.com/moby/buildkit/util/appcontext"
	"github.com/stretchr/testify/require"

	"github.com/myfintech/ark/src/go/lib/container"
)

var client container.Docker
var ctx = appcontext.Context()

func init() {
	docker, err := container.NewDockerClient(container.DefaultDockerCLIOptions()...)
	if err != nil {
		panic(err)
	}
	client = *docker
}

func TestAction(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	testdata := filepath.Join(cwd, "testdata")

	target := &Target{
		Repo:       "gcr.io/[insert-google-project]/ark/test",
		Dockerfile: "FROM node",
		RawTarget: ark.RawTarget{
			Name:  "example",
			Realm: cwd,
			Type:  Type,
			File:  filepath.Join(cwd, "targets_test.go"),
			SourceFiles: []string{
				filepath.Join(testdata, "01_dont_change_me.txt"),
				filepath.Join(testdata, "02_dont_change_me.txt"),
			},
		},
	}

	err = target.Validate()
	require.NoError(t, err)

	checksum, err := target.Checksum()
	require.NoError(t, err)

	artifact, err := target.Produce(checksum)
	require.NoError(t, err)

	action := &Action{
		Client:   client,
		Artifact: artifact.(*Artifact),
		Target:   target,
	}

	err = action.Execute(ctx)
	require.NoError(t, err)

	image := artifact.(*Artifact)
	image.Client = client
	require.True(t, image.Cacheable())

	locallyCached, err := image.LocallyCached(ctx)
	require.NoError(t, err)
	require.True(t, locallyCached)

	err = image.Push(ctx)
	require.NoError(t, err)

	remotelyCached, err := image.RemotelyCached(ctx)
	require.NoError(t, err)
	require.True(t, remotelyCached)
}

var decryptedSecretData = `{
 "bar": "baz",
 "foo": "bar"
}`

func TestCreateSecretSpecs(t *testing.T) {
	path := "secret/foo"

	cwd, err := os.Getwd()
	require.NoError(t, err)

	vClient, cleanup := vault_test_harness.CreateVaultTestCore(t, false)
	defer cleanup()

	storage := &kv.VaultStorage{
		Client:        vClient,
		FSBasePath:    filepath.Join(cwd, "testdata", ".ark/kv"),
		EncryptionKey: "domain-key",
	}

	defer func() {
		_ = os.RemoveAll(storage.FSBasePath)
	}()

	_, err = storage.Put(path, map[string]interface{}{
		"foo": "bar",
		"bar": "baz",
	})
	require.NoError(t, err)

	action := &Action{
		KVStore: storage,
	}

	specs, err := action.createSecretSpecs([]string{path})
	require.NoError(t, err)
	fileContent, err := os.ReadFile(specs[0].FilePath)
	require.NoError(t, err)
	require.Equal(t, decryptedSecretData, string(fileContent))

	require.NoError(t, action.removeSecretSpecs(specs))
	require.NoFileExists(t, specs[0].FilePath)
}
