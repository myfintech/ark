package test

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/streaming"

	"github.com/myfintech/ark/src/go/lib/kube"

	gonanoid "github.com/matoous/go-nanoid"
	"github.com/myfintech/ark/src/go/lib/kube/objects"
	corev1 "k8s.io/api/core/v1"
)

// Action is the executor for implementing a test
type Action struct {
	Artifact    *Artifact
	Target      *Target
	ManifestDir string
	K8sClient   kube.Client
	Logger      logz.FieldLogger
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

// Execute runs the action and produces a test.Artifact
func (a Action) Execute(ctx context.Context) (err error) {
	if a.ManifestDir == "" {
		a.ManifestDir, err = a.Artifact.CacheDirPath()
		if err != nil {
			return
		}
	}

	manifest, name := a.buildJobManifest()

	manifestItems := manifest.CoreList.Items

	if err = os.MkdirAll(filepath.Dir(a.renderedFilePath()), 0755); err != nil {
		return err
	}

	dest, err := os.OpenFile(a.renderedFilePath(), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
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

	client := a.K8sClient
	namespace := client.Namespace()
	clientSet, err := client.Factory.KubernetesClientSet()
	if err != nil {
		return err
	}

	if err = kube.Apply(client, namespace, a.renderedFilePath()); err != nil {
		return err
	}

	defer func() {
		time.Sleep(3 * time.Second) // Things are getting deleted too quickly ... apparently ...
		if a.Target.DisableCleanup {
			return
		}
		_ = kube.Delete(client, namespace, 10, a.renderedFilePath())
	}()

	startLogStream := make(chan bool)

	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		if <-startLogStream {
			return streamLogs(egCtx, client, namespace, name)
		}
		return nil
	})
	eg.Go(func() error {
		defer close(startLogStream)
		return watchJob(egCtx, clientSet, namespace, name, startLogStream)
	})

	if err = eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (a Action) renderedFilePath() string {
	return filepath.Join(a.ManifestDir, "manifest.yaml")
}

func (a Action) buildJobManifest() (*objects.Manifest, string) {
	manifest := new(objects.Manifest)
	containerEnv := make([]corev1.EnvVar, 0)

	for k, v := range a.Target.Environment {
		containerEnv = append(containerEnv, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	name := gonanoid.MustGenerate("0123456789abcdefghijklmnopqrstuvwxyz", 12)

	container := objects.Container(objects.ContainerOptions{
		Name:       a.Target.Name,
		Image:      a.Target.Image,
		Command:    a.Target.Command,
		Env:        containerEnv,
		Args:       a.Target.Args,
		WorkingDir: a.Target.WorkingDirectory,
	})

	podTemplate := objects.PodTemplate(objects.PodTemplateOptions{
		Name:          a.Target.Name,
		Containers:    []corev1.Container{*container},
		RestartPolicy: "Never",
	})

	backoffLimit := int32(0) // this should not retry on any kind of failure

	timeout := int64(300)
	if a.Target.TimeoutSeconds != 0 {
		timeout = int64(a.Target.TimeoutSeconds)
	}

	job := objects.Job(objects.JobOptions{
		Name: name,
		Labels: map[string]string{
			"ark.target.key": a.Target.KeyHash(),
		},
		PodTemplate:           *podTemplate,
		BackoffLimit:          &backoffLimit,
		ActiveDeadlineSeconds: &timeout,
	})

	manifest.Append(job)

	return manifest, name
}

// streamLogs waits for a pod phase other than 'pending' or 'unknown' and then streams logs back to the CLI
func streamLogs(ctx context.Context, client kube.Client, namespace, jobName string) error {
	clientSet, err := client.Factory.KubernetesClientSet()
	if err != nil {
		return err
	}
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
		var isPod bool
		if pod, isPod = event.Object.(*corev1.Pod); !isPod {
			continue
		}

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

	if _, err = io.Copy(client.OutputWriter, stream); err != nil {
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
