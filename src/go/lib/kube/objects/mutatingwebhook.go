package objects

import (
	adregv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WebhookOptions represents fields that can be passed in to create a mutating webhook
type WebhookOptions struct {
	Name         string
	Labels       map[string]string
	ClientConfig adregv1beta1.WebhookClientConfig
}

// MutatingWebhook returns a pointer to a mutating webhook
