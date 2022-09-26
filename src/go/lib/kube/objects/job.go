package objects

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// JobOptions represents fields that can be passed in to create a job
type JobOptions struct {
	Name                  string
	Labels                map[string]string
	Annotations           map[string]string
	PodTemplate           corev1.PodTemplateSpec
	BackoffLimit          *int32
	ActiveDeadlineSeconds *int64
}

// Job returns a pointer to a job object
func Job(opts JobOptions) *batchv1.Job {
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        opts.Name,
			Labels:      opts.Labels,
			Annotations: opts.Annotations,
		},
		Spec: batchv1.JobSpec{
			Template:              opts.PodTemplate,
			BackoffLimit:          opts.BackoffLimit,
			ActiveDeadlineSeconds: opts.ActiveDeadlineSeconds,
		},
	}
}
