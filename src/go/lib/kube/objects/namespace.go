package objects

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceOptions represents fields that can be passed in to create a namespace
type NamespaceOptions struct {
	Name   string
	Labels map[string]string
}

// Namespace returns a pointer to a namespace object
