package objects

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeploymentOptions represents fields that can be passed in to create a deployment
type DeploymentOptions struct {
	Name        string
	Replicas    *int32
	PodTemplate corev1.PodTemplateSpec
	Annotations map[string]string
}

// Deployment returns a pointer to a deployment object
func Deployment(opts DeploymentOptions) *appsv1.Deployment {
	defaultLabels := map[string]string{"app": opts.Name}
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        opts.Name,
			Labels:      defaultLabels,
			Annotations: opts.Annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: opts.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: defaultLabels,
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: "RollingUpdate",
			},
			Template: opts.PodTemplate,
		},
	}
}
