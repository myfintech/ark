package kube

import (
	"net/http"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	"github.com/pkg/errors"
)

// ForwardingOptions a container for options and channels that help orchestrate the port forwarder
type ForwardingOptions struct {
	Namespace    string
	Pod          v1.Pod
	Client       Client
	Ports        []string
	StopChannel  chan struct{}
	ReadyChannel chan struct{}
	DoneChannel  chan error

	stopped bool
}

func (f *ForwardingOptions) Stop() {
	if !f.stopped {
		close(f.StopChannel)
	}
	f.stopped = true
}

// PortForward builds out all of the required options to port forward, aggregates them into the ForwardingOptions struct
// and then runs the Execute() function to run the port forward operation
func PortForward(opts ForwardingOptions) (err error) {
	restClient, err := opts.Client.Factory.RESTClient()
	if err != nil {
		return err
	}

	restConfig, err := opts.Client.Factory.ToRESTConfig()
	if err != nil {
		return err
	}

	if opts.Pod.Status.Phase != v1.PodRunning {
		return errors.Errorf("unable to forward port because pod is not running. Current status=%v",
			opts.Pod.Status.Phase)
	}

	req := restClient.Post().
		Resource("pods").
		Namespace(opts.Namespace).
		Name(opts.Pod.Name).
		SubResource("portforward")

	transport, upgrader, err := spdy.RoundTripperFor(restConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())
	fw, err := portforward.New(
		dialer,
		opts.Ports,
		opts.StopChannel,
		opts.ReadyChannel,
		opts.Client.OutputWriter,
		opts.Client.OutputWriter,
	)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}
