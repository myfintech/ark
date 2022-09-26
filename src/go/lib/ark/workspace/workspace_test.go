package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDetermineRoot(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	testdata := filepath.Join(cwd, "testdata")

	t.Run("should determine root from testdata directory", func(t *testing.T) {
		root, err := DetermineRoot(testdata)
		require.NoError(t, err)
		require.Equal(t, testdata, root)
	})

	t.Run("should determine root from deeply nested directory", func(t *testing.T) {
		root, err := DetermineRoot(filepath.Join(testdata, "test_level_1/test_level_2/test_level_3"))
		require.NoError(t, err)
		require.Equal(t, testdata, root)
	})

	t.Run("should fail on attempting to determine root when it does not exist", func(t *testing.T) {
		_, err := DetermineRoot("/tmp")
		require.Error(t, err, "test should fail from tmp dir")
	})
}

func TestDetermineRootFromCWD(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		_ = os.Chdir(cwd)
	}()

	testdata := filepath.Join(cwd, "testdata")

	err = os.Chdir(filepath.Join(testdata, "test_level_1/test_level_2/test_level_3"))
	require.NoError(t, err)

	t.Run("should determine root from cwd directory", func(t *testing.T) {
		root, err := DetermineRootFromCWD()
		require.NoError(t, err)
		require.Equal(t, testdata, root)
	})
}

func TestLoadConfig(t *testing.T) {
	t.Run("should load settings from json file", func(t *testing.T) {
		cwd, err := os.Getwd()
		require.NoError(t, err)

		testdata := filepath.Join(cwd, "testdata")

		config, err := LoadConfig(testdata)
		require.NoError(t, err)
		require.Equal(t, mockConfig(testdata), config)
		require.Equal(t, configFilePath(testdata), config.File())
		require.Equal(t, filepath.Join(testdata, ".ark"), config.Dir())
		require.Equal(t, testdata, config.Root())
	})
}

func mockConfig(testdata string) *Config {
	return &Config{
		file: configFilePath(testdata),
		K8s: KubernetesConfig{
			SafeContexts: []string{
				"development_sre",
			},
			Namespace: "test-namespace",
		},
		Vault: VaultConfig{
			Address:       "https://vault.mantl.team",
			EncryptionKey: "mantl-key",
		},
		FileSystem: FileSystemConfig{Ignore: []string{"test"}},
		Plugins: []Plugin{{
			Name:  "test",
			Image: "gcr.io",
		}},
		ControlPlane:         ControlPlaneConfig{},
		User:                 UserConfig{},
		Internal:             InternalConfig{},
		VersionCheckDisabled: true,
	}
}

func TestLoadConfigFromCWD(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	defer func() {
		_ = os.Chdir(cwd)
	}()

	testdata := filepath.Join(cwd, "testdata")

	err = os.Chdir(filepath.Join(testdata, "test_level_1/test_level_2/test_level_3"))
	require.NoError(t, err)

	t.Run("should determine root from cwd directory", func(t *testing.T) {
		config, err := LoadConfigFromCWD()
		require.NoError(t, err)
		require.Equal(t, mockConfig(testdata), config)
		require.Equal(t, configFilePath(testdata), config.File())
		require.Equal(t, filepath.Join(testdata, ".ark"), config.Dir())
		require.Equal(t, testdata, config.Root())
	})
}

func TestKubernetesNamespaceOverride(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	testdata := filepath.Join(cwd, "testdata")
	t.Run("should contain Kubernetes namespace parameter", func(t *testing.T) {
		config, err := LoadConfig(testdata)
		require.NoError(t, err)
		require.Equal(t, "test-namespace", config.K8s.Namespace)
	})
}
