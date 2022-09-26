package pkg

import (
	"bytes"

	"github.com/myfintech/ark/src/go/lib/kube/objects"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type DatadogOptions struct {
	Image           string            `json:"image"`
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`
	ArkVersion      string            `json:"arkVersion"`
	EnableLogging   bool              `json:"enableLogging"`
	EnableAPM       bool              `json:"enableApm"`
	LoggingInclude  string            `json:"loggingInclude"`
	LoggingExclude  string            `json:"loggingExclude"`
}

func NewDatadogManifest(opts DatadogOptions) (string, error) {
	manifest := new(objects.Manifest)
	buff := new(bytes.Buffer)

	serviceAccount := objects.ServiceAccount(objects.ServiceAccountOptions{
		Name: "datadog-agent",
	})

	podTemplate := objects.PodTemplate(objects.PodTemplateOptions{
		Name: "datadog-agent",
		Containers: []corev1.Container{
			*buildAgentContainer(&opts),
			*buildProcessAgentContainer(&opts),
			*buildSystemProbeContainer(&opts),
		},
		NodeSelector: map[string]string{
			"kubernetes.io/os": "linux",
		},
		ServiceAccountName: "datadog-agent",
		InitContainers: []corev1.Container{
			*buildInitVolumeContainer(&opts),
			*buildInitConfigContainer(&opts),
			*buildSeccompSetupContainer(&opts),
		},
		Volumes: getDatadogPodTemplateVolumes(&opts),
		Annotations: map[string]string{
			"container.apparmor.security.beta.kubernetes.io/system-probe": "unconfined",
			"container.seccomp.security.alpha.kubernetes.io/system-probe": "localhost/system-probe",
		},
		Labels: map[string]string{
			"app": "datadog-agent",
		},
	})

	daemonSet := objects.DaemonSet(objects.DaemonSetOptions{
		Name:        "datadog-agent",
		PodTemplate: *podTemplate,
		UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
			Type:          appsv1.RollingUpdateDaemonSetStrategyType,
			RollingUpdate: &appsv1.RollingUpdateDaemonSet{MaxUnavailable: &intstr.IntOrString{Type: intstr.String, StrVal: "10%"}},
		},
	})

	if opts.EnableAPM {
		daemonSet.Spec.Template.Spec.Containers = append(daemonSet.Spec.Template.Spec.Containers, *buildAPMContainer(&opts))
	}

	manifest.Append(
		buildInstallInfoConfigMap(&opts),
		buildSystemProbeConfigMap(),
		buildSecurityConfigMap(),
		serviceAccount,
		daemonSet,
	)

	if err := manifest.Serialize(buff); err != nil {
		return "", err
	}
	return buff.String(), nil
}
