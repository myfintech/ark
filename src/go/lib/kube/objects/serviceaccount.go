package objects

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceAccountOptions represents fields that can be passed in to create a service account
type ServiceAccountOptions struct {
	Name   string
	Labels map[string]string
}

// ServiceAccount returns a pointer to a service account
func ServiceAccount(opts ServiceAccountOptions) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   opts.Name,
			Labels: opts.Labels,
		},
	}
}
