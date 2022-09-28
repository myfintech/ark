package test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/kube"
	"github.com/myfintech/ark/src/go/lib/utils"

	"github.com/stretchr/testify/require"
)

var (
	k8sClient    kube.Client
	namespace    = utils.EnvLookup("ARK_K8S_NAMESPACE", "default")
	safeContexts = utils.EnvLookup("ARK_K8S_SAFE_CONTEXTS", "")
)

func init() {
	var sContexts []string
	sContexts = append(sContexts, kube.DefaultSafeContexts()...)
	sContexts = append(sContexts, strings.Split(safeContexts, ",")...)
	k8s, err := kube.InitWithSafeContexts(namespace, sContexts)
	if err != nil {
		panic(err)
	}
	k8sClient = k8s
}

func TestTest(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	testdata := filepath.Join(cwd, "testdata")

	tests := []struct {
		wantsErr bool
		target   Target
	}{
		{
			target: Target{
				RawTarget: ark.RawTarget{
					Type:  Type,
					Name:  "example-success",
					File:  filepath.Join(cwd, "testing_targets.go"),
					Realm: cwd,
				},
				Command: []string{"bash", "-c"},
				Args:    []string{"ls -lah | grep -o 'lib' && [[ $PWD == '/usr' ]] && [[ $$TEST == foo ]] && echo yes || echo fail"},
				Image:   "bash:latest",
				Environment: map[string]string{
					"TEST": "foo",
				},
				WorkingDirectory: "/usr",
				TimeoutSeconds:   5,
			},
		},
		{
			wantsErr: true,
			target: Target{
				RawTarget: ark.RawTarget{
					Type:  Type,
					Name:  "example-fail",
					File:  filepath.Join(cwd, "testing_targets.go"),
					Realm: cwd,
				},
				Command:        []string{"bash", "-c"},
				Args:           []string{"ls -lah | grep -o 'fail'"},
				Image:          "bash:latest",
				TimeoutSeconds: 5,
			},
		},
	}

	for _, test := range tests {
		require.NoError(t, test.target.Validate())

		action := &Action{
			Target:      &test.target,
			ManifestDir: testdata,
			K8sClient:   k8sClient,
		}
		require.Implements(t, (*ark.Action)(nil), action)
		exErr := action.Execute(context.Background())
		if test.wantsErr {
			require.Error(t, exErr)
		} else {
			require.NoError(t, exErr)
		}
	}
}
