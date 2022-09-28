package objects

import (
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CronJobOptions represents fields that can be passed in to create a cron job
type CronJobOptions struct {
	Name              string
	Labels            map[string]string
	Annotations       map[string]string
	JobTemplate       batchv1beta1.JobTemplateSpec
	ConcurrencyPolicy batchv1beta1.ConcurrencyPolicy
}

// CronJob returns a pointer to a cron job object
func CronJob(opts CronJobOptions) *batchv1beta1.CronJob {
	if opts.ConcurrencyPolicy == "" {
		opts.ConcurrencyPolicy = "Allow"
	}
	return &batchv1beta1.CronJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CronJob",
			APIVersion: "batch/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        opts.Name,
			Labels:      opts.Labels,
			Annotations: opts.Annotations,
		},
		Spec: batchv1beta1.CronJobSpec{
			JobTemplate:       opts.JobTemplate,
			ConcurrencyPolicy: opts.ConcurrencyPolicy,
		},
	}
}
