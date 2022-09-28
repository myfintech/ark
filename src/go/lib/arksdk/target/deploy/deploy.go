package deploy

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/myfintech/ark/src/go/lib/ark/components/entrypoint"

	"github.com/myfintech/ark/src/go/lib/pattern"

	appsV1 "k8s.io/api/apps/v1"

	"k8s.io/apimachinery/pkg/util/yaml"

	"k8s.io/client-go/kubernetes/scheme"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/streaming"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/sync/errgroup"
	coreV1 "k8s.io/api/core/v1"

	"github.com/myfintech/ark/src/go/lib/arksdk/target/base"
	"github.com/myfintech/ark/src/go/lib/hclutils"
	"github.com/myfintech/ark/src/go/lib/kube"
	"github.com/myfintech/ark/src/go/lib/log"
)

// Target defines the required and optional attributes for defining a jsonnetutils execution
type Target struct {
	*base.RawTarget `json:"-"`

	Manifest            hcl.Expression `hcl:"manifest,attr" json:"manifest"`
	PortForward         hcl.Expression `hcl:"port_forward,optional" json:"port_forward"`
	LiveSyncEnabled     hcl.Expression `hcl:"live_sync_enabled,optional" json:"live_sync_enabled"`
	LiveSyncRestartMode hcl.Expression `hcl:"live_sync_restart_mode,optional" json:"live_sync_restart_mode"`
	LiveSyncOnActions   hcl.Expression `hcl:"live_sync_on_actions,optional" json:"live_sync_on_actions"`
	Env                 hcl.Expression `hcl:"env,optional" json:"env"`
}

// ComputedAttrs used to store the computed attributes of a local_exec target
type ComputedAttrs struct {
	Manifest            string              `hcl:"manifest,attr" json:"manifest"`
	PortForward         []string            `hcl:"port_forward,optional" json:"port_forward"` // TODO: There's no input validation for this in this code
	LiveSyncEnabled     bool                `hcl:"live_sync_enabled,optional" json:"live_sync_enabled"`
	LiveSyncRestartMode string              `hcl:"live_sync_restart_mode,optional" json:"live_sync_restart_mode"` // TODO: There's no input validation for this code or setting of a default
	LiveSyncOnActions   []Action            `hcl:"live_sync_on_actions,optional" json:"live_sync_on_actions"`
	Env                 []map[string]string `hcl:"env,optional" json:"env"`
}

// Action defines the required fields to execute one or more commands when a changed file matches a given pattern
type Action struct {
	Command  []string `hcl:"command,attr" cty:"command"`
	WorkDir  string   `hcl:"work_dir,attr" cty:"work_dir"`
	Patterns []string `hcl:"patterns,attr" cty:"patterns"`
}

// Attributes return combined rawTarget.Attributes with typedTarget.Attributes.
func (t Target) Attributes() map[string]cty.Value {
	return hclutils.MergeMapStringCtyValue(t.RawTarget.Attributes(), map[string]cty.Value{
		"rendered_file": cty.StringVal(t.RenderedFilePath()),
	})
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

// CacheEnabled overrides the default target caching behavior
func (t Target) CacheEnabled() {
	return
}

// PreBuild a lifecycle hook for calculating state before the build
func (t Target) PreBuild() error {
	clientSet, err := t.Workspace.K8s.Factory.KubernetesClientSet()
	if err != nil {
		return err
	}
	if _, lookupTestNamespaceErr := clientSet.CoreV1().Namespaces().Get(t.Workspace.Context, t.Workspace.K8s.Namespace(), metav1.GetOptions{}); lookupTestNamespaceErr != nil {
		if k8sErrors.IsNotFound(lookupTestNamespaceErr) {
			if _, createTestNamespaceErr := clientSet.CoreV1().Namespaces().Create(
				t.Workspace.Context,
				&coreV1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: t.Workspace.K8s.Namespace(),
					},
				},
				metav1.CreateOptions{},
			); createTestNamespaceErr != nil {
				return createTestNamespaceErr
			}
			log.Infof("Created namespace: %s", t.Workspace.K8s.Namespace())
		} else {
			return lookupTestNamespaceErr
		}
	}

	return hclutils.DecodeExpressions(&t, t.ComputedAttrs(), base.CreateEvalContext(base.EvalContextOptions{
		CurrentTarget:     t,
		Package:           *t.Package,
		TargetLookupTable: t.Workspace.TargetLUT,
		Workspace:         *t.Workspace,
	}))
}

// Build constructs a jsonnet manifest from the information provided in the jsonnet target
func (t Target) Build() error {
	attrs := t.ComputedAttrs()

	if err := t.MkArtifactsDir(); err != nil {
		return err
	}

	dest, err := os.Create(t.RenderedFilePath())
	if err != nil {
		return err
	}

	defer func() {
		_ = dest.Close()
	}()

	deserializedManifest, err := deserializeManifestString(attrs.Manifest)
	if err != nil {
		return err
	}

	for _, object := range deserializedManifest {
		t.applyLabels(object)
		if deployment, ok := object.(*appsV1.Deployment); ok {
			t.applyEnvToDeployment(deployment)
		}
	}

	manifest, err := os.OpenFile(t.RenderedFilePath(), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
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

	client := t.Workspace.K8s
	namespace := client.Namespace()

	if applyErr := kube.Apply(client, namespace, t.RenderedFilePath()); applyErr != nil {
		return applyErr
	}

	deployedResources, err := kube.GetObservableResourceNamesByLabel(client, namespace, "ark.target.address", t.Address())
	if err != nil {
		return err
	}

	eg, _ := errgroup.WithContext(t.Workspace.Context)
	for _, deployedResource := range deployedResources {
		eg.Go(func() error {
			return kube.RolloutStatus(client, namespace, 900*time.Second, deployedResource.Kind, deployedResource.Name)
		})
	}

	if err = eg.Wait(); err != nil {
		return err
	}

	// portMap, pmErr := portbinder.PortMapFromSlice(attrs.PortForward)
	// if pmErr != nil {
	// 	return pmErr
	// }
	//
	// if attrs.LiveSyncEnabled {
	// 	port, portErr := utils.GetFreePort()
	// 	if portErr != nil {
	// 		return portErr
	// 	}
	// 	portMap["ark_grpc_entrypoint"] = portbinder.Binding{
	// 		HostPort:   port,
	// 		RemotePort: "9000",
	// 	}
	// }
	//
	// if len(attrs.PortForward) > 0 || attrs.LiveSyncEnabled {
	// 	bindPortCmd := portbinder.BindPortCommand{
	// 		Selector: portbinder.Selector{
	// 			Namespace:  t.Workspace.K8s.Namespace(),
	// 			Type:       "pod",
	// 			LabelKey:   "ark.target.address",
	// 			LabelValue: t.Address(),
	// 		},
	// 		PortMap: portMap,
	// 	}
	//
	// 	t.Workspace.PortBinderCommands <- bindPortCmd
	// }

	return nil
}

// RenderedFilePath path of ark artifacts directory where the manifest will be written to
func (t Target) RenderedFilePath() string {
	return filepath.Join(t.ArtifactsDir(), "manifest.yaml")
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

func (t Target) applyLabels(object runtime.Object) {
	attrs := t.ComputedAttrs()

	var labels = map[string]string{
		"ark.target.address":    t.Address(),
		"ark.live.sync.enabled": strconv.FormatBool(attrs.LiveSyncEnabled),
	}

	switch obj := object.(type) {
	case *appsV1.Deployment:
		if obj.GetLabels() == nil {
			obj.SetLabels(labels)
		} else {
			obj.Labels["ark.target.address"] = t.Address()
			obj.Labels["ark.live.sync.enabled"] = strconv.FormatBool(attrs.LiveSyncEnabled)
			obj.Spec.Template.Labels["ark.target.address"] = t.Address()
			obj.Spec.Template.Labels["ark.live.sync.enabled"] = strconv.FormatBool(attrs.LiveSyncEnabled)
		}
	case *appsV1.DaemonSet:
		if obj.GetLabels() == nil {
			obj.SetLabels(labels)
		} else {
			obj.Labels["ark.target.address"] = t.Address()
			obj.Labels["ark.live.sync.enabled"] = strconv.FormatBool(attrs.LiveSyncEnabled)
			obj.Spec.Template.Labels["ark.target.address"] = t.Address()
			obj.Spec.Template.Labels["ark.live.sync.enabled"] = strconv.FormatBool(attrs.LiveSyncEnabled)
		}
	case *appsV1.StatefulSet:
		if obj.GetLabels() == nil {
			obj.SetLabels(labels)
		} else {
			obj.Labels["ark.target.address"] = t.Address()
			obj.Labels["ark.live.sync.enabled"] = strconv.FormatBool(attrs.LiveSyncEnabled)
			obj.Spec.Template.Labels["ark.target.address"] = t.Address()
			obj.Spec.Template.Labels["ark.live.sync.enabled"] = strconv.FormatBool(attrs.LiveSyncEnabled)
		}
	case *coreV1.Service:
		if obj.GetLabels() == nil {
			obj.SetLabels(labels)
		} else {
			obj.Labels["ark.target.address"] = t.Address()
			obj.Labels["ark.live.sync.enabled"] = strconv.FormatBool(attrs.LiveSyncEnabled)
		}
	}
}

func (t Target) applyEnvToDeployment(deployment *appsV1.Deployment) *appsV1.Deployment {
	if t.Workspace.Config.ControlPlane == nil {
		return deployment
	}

	deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, coreV1.Container{
		Env: []coreV1.EnvVar{
			{
				Name:  "ARK_EVENT_SINK_URL",
				Value: t.Workspace.Config.ControlPlane.EventSinkURL,
			},
			{
				Name:  "ARK_LOG_SINK_URL",
				Value: t.Workspace.Config.ControlPlane.LogSinkURL,
			},
			{
				Name:  "ARK_USER_TOKEN",
				Value: t.Workspace.Config.User.Token,
			},
			{
				Name:  "ARK_ORG_ID",
				Value: t.Workspace.Config.ControlPlane.OrgID,
			},
			{
				Name:  "ARK_PROJECT_ID",
				Value: t.Workspace.Config.ControlPlane.ProjectID,
			},
		},
	})

	return deployment
}

// ActionsToSend builds the pattern list for a matcher and then compares a file path against the matcher, returning a boolean and an error if there is one
func (t Target) ActionsToSend(filename string) ([]*entrypoint.Action, error) {
	attrs := t.ComputedAttrs()
	actionsToTake := make([]*entrypoint.Action, 0)
	for _, action := range attrs.LiveSyncOnActions {
		matcher := pattern.Matcher{}
		matcher.Includes = append(matcher.Includes, action.Patterns...)
		if err := matcher.Compile(); err != nil {
			return actionsToTake, err
		}
		if matcher.Check(filename) {
			syncAction := entrypoint.Action{
				Command:  action.Command,
				Workdir:  action.WorkDir,
				Patterns: action.Patterns,
			}
			actionsToTake = append(actionsToTake, &syncAction)
		}
	}
	return actionsToTake, nil
}
