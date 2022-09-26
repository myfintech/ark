package kube

import (
	"os"
	"time"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/kubectl/pkg/cmd/delete"
)

// Delete removes Kubernetes resources from a cluster based on a set of provided manifests
func Delete(client Client, namespace string, timeout time.Duration, files ...string) error {
	filenameOpts := &resource.FilenameOptions{
		Filenames: files,
	}

	mapper, err := client.Factory.ToRESTMapper()
	if err != nil {
		return err
	}
	dynamicClient, err := client.Factory.DynamicClient()
	if err != nil {
		return err
	}
	result := client.Factory.NewBuilder().
		Unstructured().
		ContinueOnError().
		NamespaceParam(namespace).DefaultNamespace().
		FilenameParam(false, filenameOpts).
		SelectAllParam(false).
		AllNamespaces(false).
		Flatten().
		Do()

	return (&delete.DeleteOptions{
		FilenameOptions: resource.FilenameOptions{
			Filenames: files,
		},

		Cascade:         true,
		DeleteNow:       true,
		WaitForDeletion: true,

		GracePeriod: -1,
		Timeout:     timeout,

		Output: "name",

		DynamicClient: dynamicClient,
		Mapper:        mapper,
		Result:        result,
		IOStreams: genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    client.OutputWriter,
			ErrOut: client.OutputWriter,
		},
	}).RunDelete(client.Factory)
}
