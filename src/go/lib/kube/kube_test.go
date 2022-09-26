package kube

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/myfintech/ark/src/go/lib/utils"

	"golang.org/x/sync/errgroup"

	"github.com/stretchr/testify/require"
)

func TestKube(t *testing.T) {
	client := Init(nil)
	currentContext, _ := client.CurrentContext()
	if !utils.IsK8sContextSafe([]string{"docker-desktop", "development_sre"}, "ARK_K8S_SAFE_CONTEXTS", currentContext) {
		t.Skip("Skipping test because context is not designated as safe")
		return
	}

	client.NamespaceOverride = utils.EnvLookup("ARK_K8S_NAMESPACE", "default")
	namespace := client.Namespace()

	t.Run(fmt.Sprintf("apply manifest to %s namespace", namespace), func(t *testing.T) {
		require.NoError(t, Apply(client, namespace, "./testdata/manifest.yaml"))
	})
	t.Run(fmt.Sprintf("get rollout status for observable resource in %s namespace", namespace), func(t *testing.T) {
		resourceSlice, runErr := GetObservableResourceNamesByLabel(client, namespace, "ark.address", "example.deploy.target")
		require.NoError(t, runErr)
		for _, v := range resourceSlice {
			require.NoError(t, RolloutStatus(client, namespace, 120*time.Second, v.Kind, v.Name))
		}
	})
	t.Run(fmt.Sprintf("get pod name in %s namespace and port forward", namespace), func(t *testing.T) {
		pods, runErr := GetPodsByLabel(client, namespace, "ark.address", "example.deploy.target")
		require.NoError(t, runErr)
		require.NotEmpty(t, pods)
		for _, pod := range pods {
			require.Contains(t, pod.Name, "hello-world-deployment-")
			stopChannel := make(chan struct{})
			readyChannel := make(chan struct{})
			doneChannel := make(chan error)
			eg, _ := errgroup.WithContext(context.Background())
			eg.Go(func() error {
				defer close(doneChannel)
				return PortForward(ForwardingOptions{
					Namespace:    namespace,
					Pod:          pod,
					Client:       client,
					Ports:        []string{"8080:80"},
					StopChannel:  stopChannel,
					ReadyChannel: readyChannel,
				})
			})
			eg.Go(func() error {
				<-readyChannel
				resp, getErr := http.Get("http://127.0.0.1:8080")
				require.NoError(t, getErr)
				require.Equal(t, http.StatusOK, resp.StatusCode)
				close(stopChannel)
				t.Log("sent context cancellation")
				return nil
			})
			select {
			case <-doneChannel:
			case <-time.After(time.Second * 10):
				t.Error("failed to stop port forwarding withing 10 seconds")
			}
			require.NoError(t, eg.Wait())
		}
	})
	t.Run(fmt.Sprintf("delete test resources from cluster in %s namespace", namespace), func(t *testing.T) {
		require.NoError(t, Delete(client, namespace, 10*time.Second, "./testdata/manifest.yaml"))
	})
	t.Run("create secret from file", func(t *testing.T) {
		file := "./testdata/secretdata.json"
		fileData, err := ioutil.ReadFile(file)
		require.NoError(t, err)
		data := map[string]string{
			filepath.Base(file): string(fileData),
		}

		secretData, err := CreateOrUpdateSecret(client, namespace, "test-secret-1", data, nil)
		require.NoError(t, err)
		require.Equal(t, string(fileData), string(secretData.Data["secretdata.json"]))
	})
	t.Run("delete secret created from file", func(t *testing.T) {
		restClient, err := client.Factory.RESTClient()
		require.NoError(t, err)
		require.NoError(t, DeleteSecret(restClient, namespace, "test-secret-1"))
	})
	t.Run("create secret from map", func(t *testing.T) {
		data := map[string]string{
			"foo": "bar",
			"fez": "baz",
		}
		secretData, err := CreateOrUpdateSecret(client, namespace, "test-secret-2", data, nil)
		require.NoError(t, err)
		require.Equal(t, data["foo"], string(secretData.Data["foo"]))
		require.Equal(t, data["fez"], string(secretData.Data["fez"]))
		require.Empty(t, string(secretData.Data["fex"]))
	})
	t.Run("update existing secret", func(t *testing.T) {
		data := map[string]string{
			"foo": "bat",
			"fez": "bad",
			"fex": "ban",
		}
		secretData, err := CreateOrUpdateSecret(client, namespace, "test-secret-2", data, nil)
		require.NoError(t, err)
		require.Equal(t, data["foo"], string(secretData.Data["foo"]))
		require.Equal(t, data["fez"], string(secretData.Data["fez"]))
		require.Equal(t, data["fex"], string(secretData.Data["fex"]))
	})
	t.Run("delete secret created from map", func(t *testing.T) {
		restClient, err := client.Factory.RESTClient()
		require.NoError(t, err)
		require.NoError(t, DeleteSecret(restClient, namespace, "test-secret-2"))
	})
}
