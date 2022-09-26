package objects

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DaemonSetOptions represents fields that can be passed in to create a daemon set
type DaemonSetOptions struct {
	Name           string
	PodTemplate    corev1.PodTemplateSpec
	Annotations    map[string]string
	Selector       map[string]string
	Labels         map[string]string
	UpdateStrategy appsv1.DaemonSetUpdateStrategy
}

// DaemonSet returns a pointer to a daemon set object
func DaemonSet(opts DaemonSetOptions) *appsv1.DaemonSet {
	defaultLabels := map[string]string{"app": opts.Name}

	// Provide backwards compatibility to resources that called this before label and
	// selector were added to type. Could combine into one compound if statement, but
	// may cause problems if one is set and the other is not
	if opts.Labels == nil {
		opts.Labels = defaultLabels
	}
	if opts.Selector == nil {
		opts.Selector = defaultLabels
	}

	return &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        opts.Name,
			Labels:      opts.Labels,
			Annotations: opts.Annotations,
		},
		Spec: appsv1.DaemonSetSpec{
			Template: opts.PodTemplate,
			Selector: &metav1.LabelSelector{
				MatchLabels: opts.Selector,
			},
			UpdateStrategy: opts.UpdateStrategy,
		},
	}
}
