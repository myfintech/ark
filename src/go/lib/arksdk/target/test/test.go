package test

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/streaming"
	"k8s.io/client-go/kubernetes"

	gonanoid "github.com/matoous/go-nanoid"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"golang.org/x/sync/errgroup"

	"github.com/hashicorp/hcl/v2"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/hclutils"
	"github.com/myfintech/ark/src/go/lib/kube"
	"github.com/myfintech/ark/src/go/lib/kube/objects"
)

// Target an executable target, when built it runs the specified command
type Target struct {
	*base.RawTarget `json:"-"`

	Command          hcl.Expression `hcl:"command,attr"`
	Image            hcl.Expression `hcl:"image,attr"`
	Environment      hcl.Expression `hcl:"environment,optional"`
	WorkingDirectory hcl.Expression `hcl:"working_directory,optional"`
	TimeoutSeconds   hcl.Expression `hcl:"timeout,optional"`
}

// ComputedAttrs used to store the computed attributes of a local_exec target
type ComputedAttrs struct {
	Command          []string          `hcl:"command,attr"`
	Image            string            `hcl:"image,attr"`
	Environment      map[string]string `hcl:"environment,optional"`
	WorkingDirectory string            `hcl:"working_directory,optional"`
	TimeoutSeconds   int               `hcl:"timeout,optional"`
}

// Attributes return a combined map of rawTarget.Attributes and typedTarget.Attributes
func (t Target) Attributes() map[string]cty.Value {
	return hclutils.MergeMapStringCtyValue(t.RawTarget.Attributes(), map[string]cty.Value{})
}

// ComputedAttrs returns a pointer to computed attributes from the state store.
// If attributes are not in the state store it will create a new pointer and insert it into the state store.
func (t Target) ComputedAttrs() *ComputedAttrs {
	if attrs, ok := t.GetStateAttrs().(*ComputedAttrs); ok {
		return attrs
	}

	attrs := &ComputedAttrs{}
	t.SetStateAttrs(attrs)
	return attrs
}

// PreBuild a lifecycle hook for calculating state before the build
func (t Target) PreBuild() error {
	return hclutils.DecodeExpressions(&t, t.ComputedAttrs(), base.CreateEvalContext(base.EvalContextOptions{
		CurrentTarget:     t,
		Package:           *t.Package,
		TargetLookupTable: t.Workspace.TargetLUT,
		Workspace:         *t.Workspace,
	}))
}

// Build executes the command specified in this target
func (t Target) Build() error {
	manifest, name := t.buildJobManifest()

	manifestItems := manifest.CoreList.Items

	if err := t.MkArtifactsDir(); err != nil {
		return err
	}

	dest, err := os.OpenFile(t.RenderedFilePath(), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer func() {
		_ = dest.Close()
	}()

	encoder := streaming.NewEncoder(dest, json.NewSerializerWithOptions(
		json.DefaultMetaFactory, nil, nil, json.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: true,
		}))

	for index, object := range manifestItems {
		if encodeErr := encoder.Encode(object.Object); encodeErr != nil {
			return encodeErr
		}
		if index != len(manifestItems)-1 {
			if _, indexErr := dest.WriteString("---\n"); indexErr != nil {
				return indexErr
			}
		}
	}

	client := t.Workspace.K8s
	namespace := client.Namespace()

	clientSet, _ := client.Factory.KubernetesClientSet()

	if err = kube.Apply(client, namespace, t.RenderedFilePath()); err != nil {
		return err
	}

	defer func() {
		_ = kube.Delete(client, namespace, 10, t.RenderedFilePath())
	}()

	startLogStream := make(chan bool)

	eg, egCtx := errgroup.WithContext(t.Workspace.Context)

	eg.Go(func() error {
		if <-startLogStream {
			return streamLogs(egCtx, clientSet, namespace, name)
		}
		return nil
	})

	eg.Go(func() error {
		defer close(startLogStream)
		return watchJob(egCtx, clientSet, namespace, name, startLogStream)
	})

	return eg.Wait()
}

// RenderedFilePath path of ark artifacts directory where the manifest will be written to
func (t Target) RenderedFilePath() string {
	return filepath.Join(t.ArtifactsDir(), "manifest.yaml")
}

// buildJobManifest uses the provided HCL attributes and returns a Kubernetes Job manifest and the generated name of the Job resource
func (t Target) buildJobManifest() (*objects.Manifest, string) {
	attrs := t.ComputedAttrs()

	manifest := new(objects.Manifest)
	containerEnv := make([]corev1.EnvVar, 0)

	for k, v := range attrs.Environment {
		containerEnv = append(containerEnv, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	name := gonanoid.MustGenerate("0123456789abcdefghijklmnopqrstuvwxyz", 12)

	container := objects.Container(objects.ContainerOptions{
		Name:       t.Name,
		Image:      attrs.Image,
		Command:    []string{"bash", "-c"},
		Env:        containerEnv,
		Args:       attrs.Command,
		WorkingDir: attrs.WorkingDirectory,
	})

	podTemplate := objects.PodTemplate(objects.PodTemplateOptions{
		Name:          t.Name,
		Containers:    []corev1.Container{*container},
		RestartPolicy: "Never",
	})

	backoffLimit := int32(0) // we don't want this to retry on any kind of failure

	timeout := int64(300)
	if attrs.TimeoutSeconds != 0 {
		timeout = int64(attrs.TimeoutSeconds)
	}

	job := objects.Job(objects.JobOptions{
		Name: name,
		Labels: map[string]string{
			"ark.target.address": t.Address(),
		},
		PodTemplate:           *podTemplate,
		BackoffLimit:          &backoffLimit,
		ActiveDeadlineSeconds: &timeout,
	})

	manifest.Append(job)

	return manifest, name
}

// streamLogs waits for a pod phase other than 'pending' or 'unknown' and then streams logs back to the CLI
func streamLogs(ctx context.Context, clientSet *kubernetes.Clientset, namespace, jobName string) error {
	podsWatcher, err := clientSet.CoreV1().Pods(namespace).Watch(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})

	defer func() {
		if podsWatcher != nil {
			podsWatcher.Stop()
		}
	}()

	if err != nil {
		return err
	}

	var pod *corev1.Pod

	for event := range podsWatcher.ResultChan() {
		pod = event.Object.(*corev1.Pod)
		if pod.Status.Phase != "Pending" && pod.Status.Phase != "Unknown" { // we want to stream logs for any running or completed state
			podsWatcher.Stop()
			break
		}
	}

	if pod == nil {
		return errors.New("Did not get a valid pod from watcher")
	}

	streamRequest := clientSet.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Follow: true,
	})

	stream, err := streamRequest.Stream(ctx)
	if err != nil {
		return err
	}

	if _, err = io.Copy(os.Stdout, stream); err != nil {
		return err
	}
	return nil
}

// watchJob observes the deployed job resource's status
// it unblocks the log streaming goroutine when the job becomes active
// it exits when the job succeeds, fails, or does something completely unforeseen
func watchJob(ctx context.Context, clientSet *kubernetes.Clientset, namespace, jobName string, logChan chan bool) error {
	watcher, err := clientSet.BatchV1().Jobs(namespace).Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", jobName),
	})

	defer func() {
		if watcher != nil {
			watcher.Stop()
		}
	}()

	if err != nil {
		return err
	}

	for event := range watcher.ResultChan() {
		job := event.Object.(*batchv1.Job)
		switch {
		case job.Status.Active > 0:
			logChan <- true
		case job.Status.Succeeded > 0:
			return nil
		case job.Status.Failed > 0:
			return errors.New("The job execution failed")
		}
	}
	return errors.Errorf("Watch completed without capturing status for %s", jobName)
}
