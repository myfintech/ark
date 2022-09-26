package pkg

import (
	"github.com/myfintech/ark/src/go/lib/kube/objects"
	corev1 "k8s.io/api/core/v1"
)

func buildProcessAgentContainer(opts *DatadogOptions) *corev1.Container {
	return objects.Container(objects.ContainerOptions{
		Name:            "process-agent",
		Image:           opts.Image,
		Command:         []string{"process-agent", "-config=/etc/datadog-agent/datadog.yaml"},
		Env:             append(getSharedEnv(), getProcessAgentEnv()...),
		VolumeMounts:    append(getSharedVolumeMounts(), getProcessAgentVolumeMounts()...),
		Resources:       corev1.ResourceRequirements{},
		ImagePullPolicy: opts.ImagePullPolicy,
	})
}

func getProcessAgentEnv() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "DD_SYSTEM_PROBE_ENABLED",
			Value: "true",
		},
		{
			Name:  "DD_ORCHESTRATOR_EXPLORER_ENABLED",
			Value: "false",
		},
		{
			Name:  "DD_PROCESS_AGENT_ENABLED",
			Value: "true",
		},
		{
			Name:  "DD_SYSTEM_PROBE_EXTERNAL",
			Value: "true",
		},
		{
			Name:  "DD_SYSPROBE_SOCKET",
			Value: "/var/run/sysprobe/sysprobe.sock",
		},
	}
}

func getProcessAgentVolumeMounts() []corev1.VolumeMount {
	mountPropagationMode := corev1.MountPropagationNone

	return []corev1.VolumeMount{
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
			Name:      "passwd",
			MountPath: "/etc/passwd",
		},
	}
}
