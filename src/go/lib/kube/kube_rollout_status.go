package kube

import (
	"os"
	"time"

	"github.com/pkg/errors"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/kubectl/pkg/cmd/rollout"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/scheme"
)

// RolloutStatus polls Kubernetes for a rollout status for Deployments, DaemonSets and StatefulSets;
// the function exits when the rollout has completed or the timeout threshold is met
func RolloutStatus(client Client, namespace string, timeout time.Duration, resourceType, resourceName string) error {
	allowedRolloutResources := []string{"Deployment", "DaemonSet", "StatefulSet"}
	builderArgs := make([]string, 0)
	dynamicClient, err := client.Factory.DynamicClient()
	if err != nil {
		return err
	}
	for _, kind := range allowedRolloutResources {
		if resourceType == kind {
			builderArgs = append(builderArgs, resourceType, resourceName)
		}
	}

	if len(builderArgs) == 0 {
		return errors.New("one of 'Deployment', 'DaemonSet', or 'StatefulSet' was not provided for rollout status")
	}

	return (&rollout.RolloutStatusOptions{
		PrintFlags:      genericclioptions.NewPrintFlags("").WithTypeSetter(scheme.Scheme),
		Namespace:       namespace,
		BuilderArgs:     builderArgs,
		Watch:           true,
		Timeout:         timeout,
		StatusViewerFn:  polymorphichelpers.StatusViewerFn,
		Builder:         client.Factory.NewBuilder,
		DynamicClient:   dynamicClient,
		FilenameOptions: &resource.FilenameOptions{},
		IOStreams: genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    client.OutputWriter,
			ErrOut: client.OutputWriter,
		},
	}).Run()
}
