package kube_exec

import (
	"context"
	"os"
	"time"

	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/myfintech/ark/src/go/lib/exec"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/myfintech/ark/src/go/lib/kube"
)

// Action is the executor for implementing a Kube Exec
type Action struct {
	Artifact  *Artifact
	Target    *Target
	K8sClient kube.Client
	Logger    logz.FieldLogger
}

var _ logz.Injector = &Action{}

// UseLogger injects a logger into the target's action
func (a *Action) UseLogger(logger logz.FieldLogger) {
	a.Logger = logger
}

// UseK8sClient injects the Kubernetes client
func (a *Action) UseK8sClient(client kube.Client) {
	a.K8sClient = client
}

// TODO: inject namespace to the k8s client

// Execute runs the action and produces a kube_exec.Artifact
func (a Action) Execute(_ context.Context) (err error) {
	timeout := 10 * time.Second
	if a.Target.TimeoutSeconds != 0 {
		timeout = time.Duration(a.Target.TimeoutSeconds) * time.Second
	}

	pod, err := a.K8sClient.GetPodByResource(
		a.Target.ResourceType,
		a.Target.ResourceName,
		timeout,
	)
	if err != nil {
		return err
	}

	if err = exec.KubernetesExecutor(exec.KubernetesExecOptions{
		Factory:       a.K8sClient.Factory,
		Pod:           pod,
		Executor:      exec.DefaultKubernetesExecutor(),
		ContainerName: a.Target.ContainerName,
		Command:       a.Target.Command,
		IOStreams: genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
		GetPodTimeout: timeout,
	}); err != nil {
		return err
	}

	return nil
}
