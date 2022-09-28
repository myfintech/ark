package kube

import (
	"os"

	"k8s.io/apimachinery/pkg/util/sets"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/kubectl/pkg/cmd/apply"
	"k8s.io/kubectl/pkg/cmd/delete"
)

// Apply deploys one or more Kubernetes manifests to a cluster
func Apply(client Client, namespace string, files ...string) error {
	recordFlags := genericclioptions.NewRecordFlags()
	recorder, err := recordFlags.ToRecorder()
	if err != nil {
		return err
	}
	printFlags := genericclioptions.NewPrintFlags("created")
	printer := func(operation string) (printers.ResourcePrinter, error) {
		printFlags.NamePrintFlags.Operation = operation
		return printFlags.ToPrinter()
	}

	dynamicClient, err := client.Factory.DynamicClient()
	if err != nil {
		return err
	}

	restMapper, err := client.Factory.ToRESTMapper()
	if err != nil {
		return err
	}

	deleteOptions := &delete.DeleteOptions{
		FilenameOptions: resource.FilenameOptions{
			Filenames: files,
		},
	}

	return (&apply.ApplyOptions{
		RecordFlags: recordFlags,
		Recorder:    recorder,

		PrintFlags: printFlags,
		ToPrinter:  printer,

		DeleteFlags:   delete.NewDeleteFlags(""),
		DeleteOptions: deleteOptions,

		Overwrite: true,

		Builder:       client.Factory.NewBuilder(),
		Mapper:        restMapper,
		DynamicClient: dynamicClient,
		Namespace:     namespace,
		IOStreams: genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    client.OutputWriter,
			ErrOut: client.OutputWriter,
		},

		VisitedUids:       sets.NewString(),
		VisitedNamespaces: sets.NewString(),
	}).Run()
}
