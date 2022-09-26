package kube_exec

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/kube"
	"github.com/stretchr/testify/require"
)

var client kube.Client

func init() {
	client = kube.Init(nil)
}

func TestKubeExec(t *testing.T) {
	// FIXME refactor this test to deploy and tear down its own resource for the test
	t.Skip("skipping test because it should not rely on a deployed resource")

	currentContext, err := client.CurrentContext()
	require.NoError(t, err)

	if currentContext != "development_sre" {
		t.Skip("please change context to 'development_sre' for this test")
	}

	cwd, err := os.Getwd()
	require.NoError(t, err)

	client.NamespaceOverride = "nginx-ingress"

	tests := []struct {
		wantErr bool
		target  Target
	}{
		{
			target: Target{
				RawTarget: ark.RawTarget{
					ID:    "test",
					Name:  "example_success",
					Type:  "test",
					File:  filepath.Join(cwd, "targets_test.go"),
					Realm: cwd,
				},
				ResourceType:   "ds",
				ResourceName:   "nginx-ingress",
				Command:        []string{"sh", "-c", "ls -lah"},
				ContainerName:  "",
				TimeoutSeconds: 10,
			},
		},
		{
			wantErr: true,
			target: Target{
				RawTarget: ark.RawTarget{
					ID:    "test",
					Name:  "example_fail",
					Type:  "test",
					File:  filepath.Join(cwd, "targets_test.go"),
					Realm: cwd,
				},
				ResourceType:   "ds",
				ResourceName:   "nginx-ingress",
				Command:        []string{"sh", "-c", "ls -lah | grep -o 'fail'"},
				ContainerName:  "",
				TimeoutSeconds: 10,
			},
		},
	}

	for _, test := range tests {
		err = test.target.Validate()
		require.NoError(t, err)

		checksum, err := test.target.Checksum()
		require.NoError(t, err)

		artifact, err := test.target.Produce(checksum)
		require.NoError(t, err)

		action := &Action{
			Target:    &test.target,
			Artifact:  artifact.(*Artifact),
			K8sClient: client,
		}

		require.Implements(t, (*ark.Action)(nil), action)

		actErr := action.Execute(context.Background())
		if test.wantErr {
			require.Error(t, actErr)
		} else {
			require.NoError(t, actErr)
		}
	}
}
