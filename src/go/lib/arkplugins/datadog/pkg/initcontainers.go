package pkg

import (
	"github.com/myfintech/ark/src/go/lib/kube/objects"
	corev1 "k8s.io/api/core/v1"
)

func buildInitVolumeContainer(opts *DatadogOptions) *corev1.Container {
	return objects.Container(objects.ContainerOptions{
		Name:    "init-volume",
		Image:   opts.Image,
		Command: []string{"bash", "-c"},
		Args:    []string{"cp -r /etc/datadog-agent /opt"},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "config",
				MountPath: "/opt/datadog-agent",
			},
		},
		ImagePullPolicy: opts.ImagePullPolicy,
	})
}

func buildInitConfigContainer(opts *DatadogOptions) *corev1.Container {
	return objects.Container(objects.ContainerOptions{
		Name:            "init-config",
		Image:           opts.Image,
		Command:         []string{"bash", "-c"},
		Env:             getSharedEnv(),
		Args:            []string{"for script in $(find /etc/cont-init.d/ -type f -name '*.sh' | sort) ; do bash $script ; done"},
		VolumeMounts:    getInitConfigVolumeMounts(),
		ImagePullPolicy: opts.ImagePullPolicy,
	})
}

func buildSeccompSetupContainer(opts *DatadogOptions) *corev1.Container {
	return objects.Container(objects.ContainerOptions{
		Name:            "seccomp-setup",
		Image:           opts.Image,
		Command:         []string{"cp", "/etc/config/system-probe-seccomp.json", "/host/var/lib/kubelet/seccomp/system-probe"},
		VolumeMounts:    getSeccompSetupVolumeMounts(),
		ImagePullPolicy: opts.ImagePullPolicy,
	})
}

func getInitConfigVolumeMounts() []corev1.VolumeMount {
	mountPropagationMode := corev1.MountPropagationNone

	return []corev1.VolumeMount{
		{
			Name:      "config",
			MountPath: "/etc/datadog-agent",
		},
		{
			Name:             "procdir",
			MountPath:        "/host/proc",
			MountPropagation: &mountPropagationMode,
			ReadOnly:         true,
		},
		{
			Name:             "runtimesocketdir",
			MountPath:        "/host/var/run",
			MountPropagation: &mountPropagationMode,
			ReadOnly:         true,
		},
		{
			Name:      "sysprobe-config",
			MountPath: "/etc/datadog-agent/system-probe.yaml",
			SubPath:   "system-probe.yaml",
		},
	}
}

func getSeccompSetupVolumeMounts() []corev1.VolumeMount {
	mountPropagationMode := corev1.MountPropagationNone

	return []corev1.VolumeMount{
		{
			Name:      "datadog-agent-security",
			MountPath: "/etc/config",
		},
		{
			Name:             "seccomp-root",
			MountPath:        "/host/var/lib/kubelet/seccomp",
			MountPropagation: &mountPropagationMode,
		},
	}
}
