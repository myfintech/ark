package exec

import (
	"io"
	"net/url"
	"time"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	restclient "k8s.io/client-go/rest"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/cmd/exec"
)

// KubernetesExecOptions for executing a command in a container in a pod
type KubernetesExecOptions struct {
	Factory       cmdutil.Factory
	Pod           *corev1.Pod
	Executor      exec.RemoteExecutor
	ContainerName string
	Command       []string
	Stdin         bool
	TTY           bool
	Quiet         bool
	IOStreams     genericclioptions.IOStreams
	GetPodTimeout time.Duration
}

// KubernetesExecutor runs a command in a container in a pod
func KubernetesExecutor(k KubernetesExecOptions) error {
	clientset, err := k.Factory.KubernetesClientSet()
	if err != nil {
		return errors.Wrap(err, "couldn't get a client set")
	}

	restConfig, err := k.Factory.ToRESTConfig()
	if err != nil {
		return errors.Wrap(err, "couldn't get a rest config")
	}

	if k.GetPodTimeout == 0 {
		k.GetPodTimeout = 5 * time.Second
	}

	return (&exec.ExecOptions{
		Command:       k.Command,
		Executor:      k.Executor,
		GetPodTimeout: k.GetPodTimeout,

		PodClient: clientset.CoreV1(),
		Config:    restConfig,

		StreamOptions: exec.StreamOptions{
			Namespace:     k.Pod.Namespace,
			PodName:       k.Pod.Name,
			ContainerName: k.ContainerName,
			Stdin:         k.Stdin,
			TTY:           k.TTY,
			Quiet:         k.Quiet,
			IOStreams:     k.IOStreams,
		},
	}).Run()
}

// DefaultKubernetesExecutor is a convenience method so we don't have to import all of the k8s packages to do this
func DefaultKubernetesExecutor() exec.RemoteExecutor {
	return &exec.DefaultRemoteExecutor{}
}

// FakeKubernetesExecutor is a convenience method for testing purposes
func FakeKubernetesExecutor() exec.RemoteExecutor {
	return &fakeRemoteExecutor{}
}

type fakeRemoteExecutor struct {
	method  string
	url     *url.URL
	execErr error
}

// Execute is a required method for the fakeRemoteExecutor struct to implement the exec.RemoteExecutor interface
func (f *fakeRemoteExecutor) Execute(method string, url *url.URL, _ *restclient.Config, _ io.Reader, _, _ io.Writer, _ bool, _ remotecommand.TerminalSizeQueue) error {
	f.method = method
	f.url = url
	return f.execErr
}
