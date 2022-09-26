package pkg

import corev1 "k8s.io/api/core/v1"

func getDatadogPodTemplateVolumes(opts *DatadogOptions) []corev1.Volume {
	hostPathType := corev1.HostPathDirectoryOrCreate

	volumes := []corev1.Volume{
		{
			Name: "installinfo",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "datadog-agent-installinfo",
					},
				},
			},
		},
		{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "runtimesocketdir",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/var/run",
				},
			},
		},
		{
			Name: "procdir",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/proc",
				},
			},
		},
		{
			Name: "cgroups",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/sys/fs/cgroup",
				},
			},
		},
		{
			Name: "sysprobe-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "datadog-agent-system-probe-config",
					},
				},
			},
		},
		{
			Name: "datadog-agent-security",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "datadog-agent-security",
					},
				},
			},
		},
		{
			Name: "seccomp-root",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/var/lib/kubelet/seccomp",
				},
			},
		},
		{
			Name: "debugfs",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/sys/kernel/debug",
				},
			},
		},
		{
			Name: "sysprobe-socket-dir",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "passwd",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/etc/passwd",
				},
			},
		},
		{
			Name: "apmsocket",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/var/run/datadog",
					Type: &hostPathType,
				},
			},
		},
	}

	if opts.EnableLogging {
		volumes = append(volumes, []corev1.Volume{
			{
				Name: "pointerdir",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/var/lib/datadog-agent/logs",
					},
				},
			},
			{
				Name: "logpodpath",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/var/log/pods",
					},
				},
			},
			{
				Name: "logdockercontainerpath",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/var/lib/docker/containers",
					},
				},
			},
		}...,
		)
	}
	return volumes
}
