package pkg

import (
	"github.com/myfintech/ark/src/go/lib/kube/objects"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func buildAPMContainer(opts *DatadogOptions) *corev1.Container {
	return objects.Container(objects.ContainerOptions{
		Name:    "tracer",
		Image:   opts.Image,
		Command: []string{"trace-agent", "-config=/etc/datadog-agent/datadog.yaml"},
		Env:     append(getSharedEnv(), getAPMEnv(opts)...),
		Ports: []corev1.ContainerPort{
			{
				Name:          "traceport",
				HostPort:      8126,
				ContainerPort: 8126,
				Protocol:      "TCP",
			},
		},
		VolumeMounts:    append(getSharedVolumeMounts(), getAPMVolumeMounts(opts)...),
		ImagePullPolicy: opts.ImagePullPolicy,
		LivenessProbe: corev1.Probe{
			Handler: corev1.Handler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(8126),
				},
			},
			InitialDelaySeconds: 15,
			TimeoutSeconds:      5,
			PeriodSeconds:       15,
		},
	})
}

func getAPMEnv(_ *DatadogOptions) []corev1.EnvVar {
	containerEnv := []corev1.EnvVar{
		{
			Name:  "DD_APM_RECEIVER_PORT",
			Value: "8126",
		},
		{
			Name:  "DD_APM_RECEIVER_SOCKET",
			Value: "/var/run/datadog/apm.socket",
		},
		{
			Name:  "DD_APM_ENABLED",
			Value: "true",
		},
		{
			Name:  "DD_APM_NON_LOCAL_TRAFFIC",
			Value: "true",
		},
	}

	return containerEnv
}

func getAPMVolumeMounts(_ *DatadogOptions) []corev1.VolumeMount {
	mountPropagationMode := corev1.MountPropagationNone

	volumeMounts := []corev1.VolumeMount{
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
			Name:      "apmsocket",
			MountPath: "/var/run/datadog",
		},
	}

	return volumeMounts
}
