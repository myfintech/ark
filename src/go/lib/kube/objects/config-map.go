package objects

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigMapOptions represents fields that can be passed in to create a config map
type ConfigMapOptions struct {
	Name   string
	Labels map[string]string
	Data   map[string]string
}

// ConfigMap returns a pointer to a config map object
func ConfigMap(opts ConfigMapOptions) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   opts.Name,
			Labels: opts.Labels,
		},
		Data: opts.Data,
	}
}
