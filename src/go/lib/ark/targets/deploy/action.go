package deploy

import (
	"bytes"
	"context"
	jsonEncoder "encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/myfintech/ark/src/go/lib/logz"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/myfintech/ark/src/go/lib/ark/cqrs"
	"github.com/myfintech/ark/src/go/lib/kube/portbinder"
	"github.com/myfintech/ark/src/go/lib/utils"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"

	"github.com/myfintech/ark/src/go/lib/kube"
	"golang.org/x/sync/errgroup"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/streaming"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
)

// Action is the executor for deploying a manifest
type Action struct {
	Target      *Target
	Artifact    *Artifact
	K8sClient   kube.Client
	Broker      cqrs.Broker
	ManifestDir string
	Logger      logz.FieldLogger
}

var _ logz.Injector = &Action{}

// UseLogger injects a logger into the target's action
func (a *Action) UseLogger(logger logz.FieldLogger) {
	a.Logger = logger
}

// UseBroker injects the broker into the Action
func (a *Action) UseBroker(client cqrs.Broker) {
	a.Broker = client
}

// UseK8sClient injects the Kubernetes client into the Action
func (a *Action) UseK8sClient(client kube.Client) {
	a.K8sClient = client
}

func (a Action) createNamespaceIfNotExists(ctx context.Context) error {
	namespace := a.K8sClient.Namespace()
	clientSet, err := a.K8sClient.Factory.KubernetesClientSet()
	if err != nil {
		return err
	}

	_, err = clientSet.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil && !k8sErrors.IsNotFound(err) {
		return err
	}

	_, err = clientSet.CoreV1().Namespaces().Create(
		ctx,
		&coreV1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: a.K8sClient.Namespace(),
			},
		},
		metav1.CreateOptions{},
	)
	if err != nil && !k8sErrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

// Execute runs the action and produces a deploy.Artifact
func (a Action) Execute(ctx context.Context) error {
	client := a.K8sClient
	namespace := client.Namespace()

	manifestDir, err := a.Artifact.MkCacheDir()
	if err != nil {
		return err
	}

	if a.ManifestDir == "" {
		// FIXME: create subdirectory that matches target
		a.ManifestDir = manifestDir
	}

	err = a.createNamespaceIfNotExists(ctx)
	if err != nil {
		return err
	}

	err = applyArkMutationsToManifest(a.renderedFilePath(), a.Target)
	if err != nil {
		return err
	}
	if err = kube.Apply(client, namespace, a.renderedFilePath()); err != nil {
		return err
	}

	deployedResources, err := kube.GetObservableResourceNamesByLabel(
		client,
		namespace,
		"ark.target.key",
		a.Target.KeyHash(),
	)
	if err != nil {
		return err
	}

	eg, _ := errgroup.WithContext(ctx)
	for _, deployedResource := range deployedResources {
		eg.Go(watchRollout(client, namespace, deployedResource))
	}

	if err = eg.Wait(); err != nil {
		return err
	}

	return nil
}

func watchRollout(
	client kube.Client,
	namespace string,
	deployedResource kube.ObservableResource,
) func() error {
	return func() error {
		return kube.RolloutStatus(
			client,
			namespace,
			900*time.Second,
			deployedResource.Kind,
			deployedResource.Name,
		)
	}
}

func (a Action) renderedFilePath() string {
	return filepath.Join(a.ManifestDir, "manifest.yaml")
}

func applyEnvToDeployment(
	deployment *appsV1.Deployment,
	controlPlaneConfig *base.ControlPlaneConfig,
	userConfig *base.UserConfig,
) *appsV1.Deployment {
	if controlPlaneConfig == nil {
		return deployment
	}

	deployment.Spec.Template.Spec.Containers = append(
		deployment.Spec.Template.Spec.Containers,
		coreV1.Container{
			Env: []coreV1.EnvVar{
				{
					Name:  "ARK_EVENT_SINK_URL",
					Value: controlPlaneConfig.EventSinkURL,
				},
				{
					Name:  "ARK_LOG_SINK_URL",
					Value: controlPlaneConfig.LogSinkURL,
				},
				{
					Name:  "ARK_USER_TOKEN",
					Value: userConfig.Token,
				},
				{
					Name:  "ARK_ORG_ID",
					Value: controlPlaneConfig.OrgID,
				},
				{
					Name:  "ARK_PROJECT_ID",
					Value: controlPlaneConfig.ProjectID,
				},
			},
		},
	)

	return deployment
}

type labeler interface {
	GetLabels() map[string]string
	SetLabels(map[string]string)
}

type annotationer interface {
	GetAnnotations() map[string]string
	SetAnnotations(map[string]string)
}

func applyLabels(object runtime.Object, target *Target) {
	pairs := [3][2]string{
		{"ark.target.key", target.KeyHash()},
		{"ark.live.sync.enabled", strconv.FormatBool(target.LiveSyncEnabled)},
		{
			"ark.port.binding.enabled",
			strconv.FormatBool(target.PortForward != nil && len(target.PortForward) > 0),
		},
	}
	labels := map[string]string{}
	for i := 0; i < len(pairs); i++ {
		labels[pairs[i][0]] = pairs[i][1]
	}

	if l, ok := object.(labeler); ok {
		if l.GetLabels() == nil {
			l.SetLabels(labels)
		}
	}

	// this is needed due to the lack of generics in golang
	switch obj := object.(type) {
	case *appsV1.Deployment:
		for key, val := range labels {
			obj.Labels[key] = val
			obj.Spec.Template.Labels[key] = val
		}
	case *appsV1.DaemonSet:
		for key, val := range labels {
			obj.Labels[key] = val
			obj.Spec.Template.Labels[key] = val
		}
	case *appsV1.StatefulSet:
		for key, val := range labels {
			obj.Labels[key] = val
			obj.Spec.Template.Labels[key] = val
		}
	case *coreV1.Service:
		for key, val := range labels {
			obj.Labels[key] = val
		}
	}
}

func applyAnnotations(object runtime.Object, target *Target) error {
	var pairs [][2]string
	portMap := target.PortForward

	if target.LiveSyncEnabled {
		port, portErr := utils.GetFreePort()
		if portErr != nil {
			return portErr
		}
		if portMap == nil {
			portMap = make(portbinder.PortMap)
		}

		portMap["ark_grpc_entrypoint"] = portbinder.Binding{
			HostPort:   port,
			RemotePort: "9000",
		}
	}

	if portMap != nil && len(portMap) > 0 {

		portForwardJsonPayload, err := jsonEncoder.Marshal(portMap)
		if err != nil {
			return err
		}

		pairs = append(
			pairs,
			[2]string{"ark.port.binding", string(portForwardJsonPayload)},
		)
	}

	annotations := map[string]string{}
	for i := 0; i < len(pairs); i++ {
		annotations[pairs[i][0]] = pairs[i][1]
	}

	if a, ok := object.(annotationer); ok {
		if a.GetAnnotations() == nil {
			a.SetAnnotations(annotations)
		}
	}
	// this is needed due to the lack of generics in golang
	switch obj := object.(type) {
	case *appsV1.Deployment:
		for key, val := range annotations {
			obj.Annotations[key] = val
			if obj.Spec.Template.Annotations == nil {
				obj.Spec.Template.Annotations = map[string]string{}
			}
			obj.Spec.Template.Annotations[key] = val
		}
	case *appsV1.DaemonSet:
		for key, val := range annotations {
			obj.Annotations[key] = val
			if obj.Spec.Template.Annotations == nil {
				obj.Spec.Template.Annotations = map[string]string{}
			}
			obj.Spec.Template.Annotations[key] = val
		}
	case *appsV1.StatefulSet:
		for key, val := range annotations {
			obj.Annotations[key] = val
			if obj.Spec.Template.Annotations == nil {
				obj.Spec.Template.Annotations = map[string]string{}
			}
			obj.Spec.Template.Annotations[key] = val
		}
	case *coreV1.Service:
		for key, val := range annotations {
			obj.Annotations[key] = val
		}
	}
	return nil
}

func deserializeManifestString(manifest string) ([]runtime.Object, error) {
	var decodedObjects []runtime.Object
	decoder := yaml.NewDocumentDecoder(ioutil.NopCloser(bytes.NewBufferString(manifest)))
	deserializer := streaming.NewDecoder(decoder, scheme.Codecs.UniversalDeserializer())

Decode:
	for {
		obj, _, err := deserializer.Decode(nil, nil)
		switch {
		case err == io.EOF:
			break Decode
		case err != nil:
			return decodedObjects, err
		}
		decodedObjects = append(decodedObjects, obj)
	}
	return decodedObjects, nil
}

func applyArkMutationsToManifest(fileName string, target *Target) error {
	if err := os.MkdirAll(filepath.Dir(fileName), 0o700); err != nil {
		return err
	}

	dest, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer func() {
		_ = dest.Close()
	}()

	deserializedManifest, err := deserializeManifestString(target.Manifest)
	if err != nil {
		return err
	}

	for _, object := range deserializedManifest {
		applyLabels(object, target)
		if err := applyAnnotations(object, target); err != nil {
			return err
		}
		if deployment, ok := object.(*appsV1.Deployment); ok {
			// FIXME inject config dependencies
			applyEnvToDeployment(deployment, nil, nil)
		}
	}

	manifest, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}

	defer func() {
		_ = manifest.Close()
	}()

	encoder := streaming.NewEncoder(manifest, json.NewSerializerWithOptions(
		json.DefaultMetaFactory, nil, nil, json.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: true,
		}))

	for index, object := range deserializedManifest {
		if encodeErr := encoder.Encode(object); encodeErr != nil {
			return encodeErr
		}
		if index != len(deserializedManifest)-1 {
			if _, indexErr := manifest.WriteString("---\n"); indexErr != nil {
				return indexErr
			}
		}
	}
	return nil
}
