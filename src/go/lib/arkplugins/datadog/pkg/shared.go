package pkg

import corev1 "k8s.io/api/core/v1"

func getSharedEnv() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: "DD_API_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "datadog-api-key",
					},
					Key: "api-key",
				},
			},
		},
		{
			Name: "DD_KUBERNETES_KUBELET_HOST",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.hostIP",
				},
			},
		},
		{
			Name:  "KUBERNETES",
			Value: "yes",
		},
		{
			Name:  "DOCKER_HOST",
			Value: "unix://host/var/run/docker.sock",
		},
		{
			Name:  "DD_LOG_LEVEL",
			Value: "INFO",
		},
	}
}

func getSharedVolumeMounts() []corev1.VolumeMount {
	mountPropagationMode := corev1.MountPropagationNone

	return []corev1.VolumeMount{
		{
			Name:      "sysprobe-socket-dir",
			MountPath: "/var/run/sysprobe",
		},
		{
			Name:      "sysprobe-config",
			MountPath: "/etc/datadog-agent/system-probe.yaml",
			SubPath:   "system-probe.yaml",
		},
		{
			Name:             "procdir",
			MountPath:        "/host/proc",
			MountPropagation: &mountPropagationMode,
			ReadOnly:         true,
		},
		{
			Name:             "cgroups",
			MountPath:        "/host/sys/fs/cgroup",
			MountPropagation: &mountPropagationMode,
			ReadOnly:         true,
		},
	}
}
