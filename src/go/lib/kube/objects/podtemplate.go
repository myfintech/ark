package objects

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodTemplateOptions represents fields that can be passed in to create a pod template spec
type PodTemplateOptions struct {
	Name                          string
	Containers                    []corev1.Container
	NodeSelector                  map[string]string
	RestartPolicy                 corev1.RestartPolicy
	ServiceAccountName            string
	InitContainers                []corev1.Container
	HostNetworkingEnabled         bool
	Volumes                       []corev1.Volume
	Annotations                   map[string]string
	Labels                        map[string]string
	TerminationGracePeriodSeconds *int64
	Affinity                      corev1.Affinity
	SecurityContext               corev1.PodSecurityContext
}

// PodTemplate returns a pointer to a pod template spec object
func PodTemplate(opts PodTemplateOptions) *corev1.PodTemplateSpec {
	if opts.RestartPolicy == "" {
		opts.RestartPolicy = "Always"
	}
	if opts.TerminationGracePeriodSeconds == nil {
		tgps := int64(30)
		opts.TerminationGracePeriodSeconds = &tgps
	}
	automountServiceAccountToken := true
	return &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        opts.Name,
			Labels:      opts.Labels,
			Annotations: opts.Annotations,
		},
		Spec: corev1.PodSpec{
			Volumes:                       opts.Volumes,
			InitContainers:                opts.InitContainers,
			Containers:                    opts.Containers,
			RestartPolicy:                 opts.RestartPolicy,
			TerminationGracePeriodSeconds: opts.TerminationGracePeriodSeconds,
			NodeSelector:                  opts.NodeSelector,
			ServiceAccountName:            opts.ServiceAccountName,
			HostNetwork:                   opts.HostNetworkingEnabled,
			Affinity:                      &opts.Affinity,
			AutomountServiceAccountToken:  &automountServiceAccountToken,
			SecurityContext:               &opts.SecurityContext,
		},
	}
}
