package base

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/myfintech/ark/src/go/lib/ark/kv"
	"github.com/myfintech/ark/src/go/lib/vault_tools/vault_test_harness"

	"github.com/zclconf/go-cty/cty/function"

	"github.com/hashicorp/hcl/v2"

	"github.com/hashicorp/hcl/v2/gohcl"

	"github.com/myfintech/ark/src/go/lib/hclutils"
	"github.com/stretchr/testify/require"
)

var loadFuncHCL = `
config = deepmerge({
  port = 3000
  vault = {
    address = "http://127.0.0.1:8200"
  }
  replicas = 2
  name = "test1"
}, load("./load_example1.hcl", { name = "test2" }, false))
`

var kvGetFuncHCL1 = `secret = kvget("test").statsdPort`
var kvGetFuncHCL2 = `secret = kvget("test").kafkaConfig.brokers`
var kvGetFuncHCL3 = `secret = kvget("test").kafkaConfig.sasl`
var kvGetFuncHCL4 = `secret = kvget("test").kafkaConfig.ssl`

func TestLoadFunc(t *testing.T) {
	rawHCL := struct {
		Config struct {
			Port     int               `cty:"port" hcl:"port,optional"`
			Vault    map[string]string `cty:"vault" hcl:"vault,optional"`
			Replicas int               `cty:"replicas" hcl:"replicas,optional"`
			Name     string            `cty:"name" hcl:"name,optional"`
			Thing    string            `cty:"thing" hcl:"thing,optional"`
		} `hcl:"config,attr"`
	}{}

	cwd, _ := os.Getwd()

	testdata := filepath.Join(cwd, "testdata_load")

	exampleHCLFile, diag := hclutils.FileFromString(loadFuncHCL)
	if diag != nil && diag.HasErrors() {
		require.NoError(t, diag)
	}
	require.NotNil(t, exampleHCLFile)

	stdlib := hclutils.BuildStdLibFunctions(testdata)
	stdlib["load"] = LoadConfigFunc(testdata)
	stdlib["deepmerge"] = hclutils.DeepMergeFunc()

	hclCtx := &hcl.EvalContext{
		Variables: nil,
		Functions: stdlib,
	}
	if exampleHCLFile != nil {
		if diag = gohcl.DecodeBody(exampleHCLFile.Body, hclCtx, &rawHCL); diag != nil && diag.HasErrors() {
			require.NoError(t, diag)
		}
	}
	require.Equal(t, "foo", rawHCL.Config.Thing)
	require.Equal(t, 3001, rawHCL.Config.Port)
	require.Equal(t, "test2", rawHCL.Config.Name)
	require.Equal(t, 2, rawHCL.Config.Replicas)
	require.Equal(t, "http://127.0.0.1:8200", rawHCL.Config.Vault["address"])
	require.Equal(t, "bar", rawHCL.Config.Vault["token"])
}

func TestKVGetFunc(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	client, cleanup := vault_test_harness.CreateVaultTestCore(t, false)
	defer cleanup()

	storage := kv.VaultStorage{
		Client:        client,
		FSBasePath:    filepath.Join(cwd, "testdata", ".ark/kv"),
		EncryptionKey: "mantl-key",
	}

	defer func() {
		_ = os.RemoveAll(storage.FSBasePath)
	}()

	_, err = storage.Put("test", map[string]interface{}{
		"kafkaConfig": map[string]interface{}{
			"brokers": []interface{}{
				"test.com:9092",
			},
			"clientId":          "consumer-lag-monitor",
			"connectionTimeout": 25000,
			"requestTimeout":    25000,
			"sasl": map[string]interface{}{
				"mechanism": "plain",
				"password":  "dummy",
				"username":  "dummy",
			},
			"ssl": true,
		},
		"statsdPort": 31825,
	})
	require.NoError(t, err)

	hclctx := &hcl.EvalContext{
		Variables: nil,
		Functions: map[string]function.Function{
			"kvget": KVGet(&storage),
		},
	}

	t.Run("bad paths", func(t *testing.T) {
		_, err := getKVData(".ark/kv/test", &storage)
		require.Error(t, err)
		require.EqualError(t, err, "the provided path should not have the base path of the KV store included")
		_, err = getKVData("thing/.ark/kv/test", &storage)
		require.Error(t, err)
		require.EqualError(t, err, "the provided path should not have the base path of the KV store included")
		_, err = getKVData("/thing/.ark/kv/test", &storage)
		require.Error(t, err)
		require.EqualError(t, err, "the provided path should not have the base path of the KV store included")
		_, err = getKVData("test/whatever", &storage)
		require.Error(t, err)
		// The error message returned in this case is specific to the host OS running the test, so we can't accurately reproduce and check the error message
	})

	t.Run("good path", func(t *testing.T) {
		data, err := getKVData("test", &storage)
		require.NoError(t, err)
		require.NotNil(t, data)
		kafkaConfig := data["kafkaConfig"].(map[string]interface{})
		require.Equal(t, "consumer-lag-monitor", kafkaConfig["clientId"])

		var brokers []string
		for _, broker := range kafkaConfig["brokers"].([]interface{}) {
			brokers = append(brokers, broker.(string))
		}
		require.Equal(t, []string{"test.com:9092"}, brokers)
		require.Equal(t, true, kafkaConfig["ssl"])
		require.Equal(t, float64(31825), data["statsdPort"])
	})

	t.Run("hcl implementation check - statsd port", func(t *testing.T) {
		rawHCL := struct {
			Secret string `hcl:"secret"`
		}{}
		exampleHCLFile, diag := hclutils.FileFromString(kvGetFuncHCL1)
		if diag != nil && diag.HasErrors() {
			require.NoError(t, diag)
		}
		require.NotNil(t, exampleHCLFile)

		if exampleHCLFile != nil {
			if diag = gohcl.DecodeBody(exampleHCLFile.Body, hclctx, &rawHCL); diag != nil && diag.HasErrors() {
				require.NoError(t, diag)
			}
		}

		require.Equal(t, "31825", rawHCL.Secret)
	})

	t.Run("hcl implementation check - brokers", func(t *testing.T) {
		rawHCL := struct {
			Secret []string `hcl:"secret"`
		}{}
		exampleHCLFile, diag := hclutils.FileFromString(kvGetFuncHCL2)
		if diag != nil && diag.HasErrors() {
			require.NoError(t, diag)
		}
		require.NotNil(t, exampleHCLFile)

		if exampleHCLFile != nil {
			if diag = gohcl.DecodeBody(exampleHCLFile.Body, hclctx, &rawHCL); diag != nil && diag.HasErrors() {
				require.NoError(t, diag)
			}
		}

		require.Equal(t, []string{"test.com:9092"}, rawHCL.Secret)
	})

	t.Run("hcl implementation check - sasl", func(t *testing.T) {
		rawHCL := struct {
			Secret map[string]string `hcl:"secret"`
		}{}
		exampleHCLFile, diag := hclutils.FileFromString(kvGetFuncHCL3)
		if diag != nil && diag.HasErrors() {
			require.NoError(t, diag)
		}
		require.NotNil(t, exampleHCLFile)

		if exampleHCLFile != nil {
			if diag = gohcl.DecodeBody(exampleHCLFile.Body, hclctx, &rawHCL); diag != nil && diag.HasErrors() {
				require.NoError(t, diag)
			}
		}

		require.Equal(t, "dummy", rawHCL.Secret["username"])
	})

	t.Run("hcl implementation check - ssl", func(t *testing.T) {
		rawHCL := struct {
			Secret bool `hcl:"secret"`
		}{}
		exampleHCLFile, diag := hclutils.FileFromString(kvGetFuncHCL4)
		if diag != nil && diag.HasErrors() {
			require.NoError(t, diag)
		}
		require.NotNil(t, exampleHCLFile)

		if exampleHCLFile != nil {
			if diag = gohcl.DecodeBody(exampleHCLFile.Body, hclctx, &rawHCL); diag != nil && diag.HasErrors() {
				require.NoError(t, diag)
			}
		}

		require.Equal(t, true, rawHCL.Secret)
	})
}
