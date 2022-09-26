package pkg

import (
	"github.com/myfintech/ark/src/go/lib/kube/objects"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func buildSystemProbeContainer(opts *DatadogOptions) *corev1.Container {
	return objects.Container(objects.ContainerOptions{
		Name:    "system-probe",
		Image:   opts.Image,
		Command: []string{"/opt/datadog-agent/embedded/bin/system-probe", "--config=/etc/datadog-agent/system-probe.yaml"},
		Env: append(getSharedEnv(), corev1.EnvVar{
			Name:  "DD_SYSPROBE_SOCKET",
			Value: "/var/run/sysprobe/sysprobe.sock",
		}),
		VolumeMounts: append(getSharedVolumeMounts(), corev1.VolumeMount{
			Name:      "debugfs",
			MountPath: "/sys/kernel/debug",
		}),
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("150Mi"),
				corev1.ResourceCPU:    resource.MustParse("200m"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("150Mi"),
				corev1.ResourceCPU:    resource.MustParse("200m"),
			},
		},
		ImagePullPolicy: opts.ImagePullPolicy,
		SecurityContext: corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{"SYS_ADMIN", "SYS_RESOURCE", "SYS_PTRACE", "NET_ADMIN", "NET_BROADCAST", "IPC_LOCK"},
			},
		},
	})
}
