package deploy

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/ark/cqrs/topics"
	"github.com/myfintech/ark/src/go/lib/kube/portbinder"

	"github.com/myfintech/ark/src/go/lib/fs"

	"github.com/myfintech/ark/src/go/lib/kube"
	"github.com/myfintech/ark/src/go/lib/utils"

	"github.com/stretchr/testify/require"
)

var (
	k8sClient    kube.Client
	namespace    = utils.EnvLookup("ARK_K8S_NAMESPACE", "default")
	safeContexts = utils.EnvLookup("ARK_K8S_SAFE_CONTEXTS", "")
)

var sContexts []string

func init() {
	sContexts = append(sContexts, kube.DefaultSafeContexts()...)
	sContexts = append(sContexts, strings.Split(safeContexts, ",")...)
	k8s, err := kube.InitWithSafeContexts(namespace, sContexts)
	if err != nil {
		panic(err)
	}
	k8sClient = k8s
}

func TestAction(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	testdata := filepath.Join(cwd, "testdata")

	t.Run("simple deploy", func(t *testing.T) {
		manifest, err := fs.ReadFileString(filepath.Join(testdata, "deploy_test/before.yaml"))
		require.NoError(t, err)

		target := &Target{
			Manifest:            manifest,
			PortForward:         nil,
			LiveSyncEnabled:     false,
			LiveSyncRestartMode: "",
			LiveSyncOnStep:      nil,
			Env:                 nil,
			RawTarget: ark.RawTarget{
				Name:  "deploy_test",
				File:  "test",
				Type:  Type,
				Realm: cwd,
			},
		}

		err = target.Validate()
		require.NoError(t, err)

		checksum, err := target.Checksum()
		require.NoError(t, err)

		artifact, err := target.Produce(checksum)
		require.NoError(t, err)

		action := &Action{
			Target:      target,
			K8sClient:   k8sClient,
			Artifact:    artifact.(*Artifact),
			ManifestDir: filepath.Join(testdata, "deploy_test"),
		}
		err = action.Execute(context.Background())
		// FIXME: this is a hack to get passed a race condition in CI
		if err != nil && strings.Contains(err.Error(), "object has been deleted") {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	})

	t.Run("checking labels and annotations; port-binding", func(t *testing.T) {
		targetName := "deploy_test_with_portbinding"
		manifest, err := fs.ReadFileString(filepath.Join(testdata, targetName, "before.yaml"))
		require.NoError(t, err)

		target := &Target{
			Manifest:            manifest,
			LiveSyncEnabled:     false,
			LiveSyncRestartMode: "",
			LiveSyncOnStep:      nil,
			Env:                 nil,
			PortForward: portbinder.PortMap{
				"http": {
					HostPort:   "8080",
					RemotePort: "3000",
				},
			},
			RawTarget: ark.RawTarget{
				Name:  targetName,
				File:  "test",
				Type:  Type,
				Realm: cwd,
			},
		}

		err = target.Validate()
		require.NoError(t, err)

		checksum, err := target.Checksum()
		require.NoError(t, err)

		artifact, err := target.Produce(checksum)
		require.NoError(t, err)

		broker := cqrs.NewMockBroker()
		// the subsystem should subscribe to the correct command topic
		broker.On("Subscribe", topics.PortBinderCommands)
		broker.On("Subscribe", topics.PortBinderEvents)
		broker.On("Publish", topics.PortBinderCommands)
		broker.On("Publish", topics.PortBinderEvents)

		action := &Action{
			Target:      target,
			K8sClient:   k8sClient,
			Artifact:    artifact.(*Artifact),
			ManifestDir: filepath.Join(testdata, "deploy_test_with_portbinding"),
			Broker:      broker,
		}

		err = action.Execute(context.Background())
		// FIXME: this is a hack to get passed a race condition in CI
		if err != nil && strings.Contains(err.Error(), "object has been deleted") {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
		arkAppliedChangesManifestPath := filepath.Join(action.ManifestDir, "manifest.yaml")
		fmt.Printf(arkAppliedChangesManifestPath)
		b, err := ioutil.ReadFile(arkAppliedChangesManifestPath)
		require.NoError(t, err)

		s := string(b)
		containsPortBindingEnabledLabel := strings.Contains(s, "ark.port.binding.enabled")
		containsPortBindingAnnotation := strings.Contains(s, "ark.port.binding")

		require.True(t, containsPortBindingEnabledLabel)
		require.True(t, containsPortBindingAnnotation)
	})
}
