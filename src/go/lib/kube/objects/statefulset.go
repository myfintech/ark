package objects

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StatefulSetOptions represents fields that can be passed in to create a stateful set
type StatefulSetOptions struct {
	Name                string
	Replicas            int32
	PodTemplate         corev1.PodTemplateSpec
	Annotations         map[string]string
	RestartPolicy       corev1.RestartPolicy
	PVCs                []corev1.PersistentVolumeClaim
	Selector            map[string]string
	Labels              map[string]string
	PodManagementPolicy appsv1.PodManagementPolicyType
}

// StatefulSet returns a pointer to a stateful set
func StatefulSet(opts StatefulSetOptions) *appsv1.StatefulSet {
	defaultLabels := map[string]string{"app": opts.Name}

	//Provide backwards compatibility to resources that called this before label and
	//selector were added to type. Could combine into one compound if statement, but
	//may cause problems if one is set and the other is not
	if opts.Labels == nil {
		opts.Labels = defaultLabels
	}
	if opts.Selector == nil {
		opts.Selector = defaultLabels
	}

	if opts.RestartPolicy == "" {
		opts.RestartPolicy = "Always"
	}

	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        opts.Name,
			Labels:      opts.Labels,
			Annotations: opts.Annotations,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &opts.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: opts.Selector,
			},
			Template:             opts.PodTemplate,
			VolumeClaimTemplates: opts.PVCs,
			ServiceName:          opts.Name,
			PodManagementPolicy:  opts.PodManagementPolicy,
		},
	}
}
