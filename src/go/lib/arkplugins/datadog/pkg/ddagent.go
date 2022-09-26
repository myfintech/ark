package pkg

import (
	"github.com/myfintech/ark/src/go/lib/kube/objects"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func buildAgentContainer(opts *DatadogOptions) *corev1.Container {
	return objects.Container(objects.ContainerOptions{
		Name:    "agent",
		Image:   opts.Image,
		Command: []string{"agent", "run"},
		Env:     append(getSharedEnv(), getAgentEnv(opts)...),
		Ports: []corev1.ContainerPort{
			{
				Name:          "dogstatsdport",
				HostPort:      31825,
				ContainerPort: 8125,
				Protocol:      "UDP",
			},
		},
		VolumeMounts:    append(getSharedVolumeMounts(), getAgentVolumeMounts(opts)...),
		ImagePullPolicy: opts.ImagePullPolicy,
		LivenessProbe: corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/live",
					Port:   intstr.IntOrString{IntVal: 5555},
					Scheme: "HTTP",
				},
			},
			InitialDelaySeconds: 15,
			TimeoutSeconds:      5,
			PeriodSeconds:       15,
			SuccessThreshold:    1,
			FailureThreshold:    6,
		},
		ReadinessProbe: corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/ready",
					Port: intstr.IntOrString{IntVal: 5555},
				},
			},
			InitialDelaySeconds: 15,
			TimeoutSeconds:      5,
			PeriodSeconds:       15,
			SuccessThreshold:    1,
			FailureThreshold:    6,
		},
	})
}

func getAgentEnv(opts *DatadogOptions) []corev1.EnvVar {
	containerEnv := []corev1.EnvVar{
		{
			Name:  "DD_DOGSTATSD_PORT",
			Value: "8125",
		},
		{
			Name:  "DD_APM_ENABLED",
			Value: "false",
		},
		{
			Name:  "DD_LOGS_CONFIG_K8S_CONTAINER_USE_FILE",
			Value: "true",
		},
		{
			Name:  "DD_HEALTH_PORT",
			Value: "5555",
		},
		{
			Name:  "DD_DOGSTATSD_NON_LOCAL_TRAFFIC",
			Value: "true",
		},
		{
			Name:  "DD_COLLECT_KUBERNETES_EVENTS",
			Value: "true",
		},
		{
			Name:  "DD_LEADER_ELECTION",
			Value: "true",
		},
	}

	if opts.EnableLogging {
		containerEnv = append(containerEnv, []corev1.EnvVar{
			{
				Name:  "DD_LOGS_ENABLED",
				Value: "true",
			},
			{
				Name:  "DD_LOGS_CONFIG_CONTAINER_COLLECT_ALL",
				Value: "true",
			}}...,
		)
	} else {
		containerEnv = append(containerEnv, []corev1.EnvVar{
			{
				Name:  "DD_LOGS_ENABLED",
				Value: "false",
			},
			{
				Name:  "DD_LOGS_CONFIG_CONTAINER_COLLECT_ALL",
				Value: "false",
			}}...,
		)
	}

	if opts.LoggingInclude != "" {
		containerEnv = append(containerEnv, corev1.EnvVar{
			Name:  "DD_CONTAINER_INCLUDE_LOGS",
			Value: "name:radius-data-warehouse image:gcr.io/managed-infrastructure/mantl/vault-auth:.*",
		})
	}

	if opts.LoggingExclude != "" {
		containerEnv = append(containerEnv, corev1.EnvVar{
			Name:  "DD_CONTAINER_EXCLUDE_LOGS",
			Value: "image:.*",
		})
	}

	return containerEnv
}

func getAgentVolumeMounts(opts *DatadogOptions) []corev1.VolumeMount {
	mountPropagationMode := corev1.MountPropagationNone

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "installinfo",
			MountPath: "/etc/datadog-agent/install_info",
			SubPath:   "install_info",
			ReadOnly:  true,
		},
		{
			Name:      "config",
			MountPath: "/etc/datadog-agent",
		},
		{
			Name:             "runtimesocketdir",
			MountPath:        "/host/var/run",
			MountPropagation: &mountPropagationMode,
			ReadOnly:         true,
		},
		{
			Name:      "debugfs",
			MountPath: "/sys/kernel/debug",
		},
	}

	if opts.EnableLogging {
		volumeMounts = append(volumeMounts, []corev1.VolumeMount{
			{
				Name:             "pointerdir",
				MountPath:        "/opt/datadog-agent/run",
				MountPropagation: &mountPropagationMode,
			},
			{
				Name:             "logpodpath",
				MountPath:        "/var/log/pods",
				MountPropagation: &mountPropagationMode,
				ReadOnly:         true,
			},
			{
				Name:             "logdockercontainerpath",
				MountPath:        "/var/lib/docker/containers",
				MountPropagation: &mountPropagationMode,
				ReadOnly:         true,
			},
		}...,
		)
	}

	return volumeMounts
}
