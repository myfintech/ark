package objects

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceOptions represents fields that can be passed in to create a service
type ServiceOptions struct {
	Name                   string
	Ports                  []corev1.ServicePort
	BackendConfig          string // example '{"ports":{"80":"nginx-backend-config"}}'
	Type                   corev1.ServiceType
	Selector               map[string]string
	Labels                 map[string]string
	Annotations            map[string]string
	LoadBalancerIP         string
	ClusterIP              string // Can be set to "None" to create a headless service
	PublishNotReadyAddress bool   // https://kubernetes.io/docs/reference/kubernetes-api/services-resources/service-v1/#ServiceSpec
}

// Service returns a pointer to a service object
func Service(opts ServiceOptions) *corev1.Service {
	annotations := make(map[string]string, 0)
	if opts.BackendConfig != "" {
		annotations["beta.cloud.google.com/backend-config"] = opts.BackendConfig
	}
	for k, v := range opts.Annotations {
		annotations[k] = v
	}

	if opts.Type == "" && opts.ClusterIP != "None" {
		opts.Type = "ClusterIP"
	}

	if opts.ClusterIP == "None" { // if using a headless service, return the ClusterIP field instead of Type
		return &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:        opts.Name,
				Labels:      opts.Labels,
				Annotations: annotations,
			},
			Spec: corev1.ServiceSpec{
				Ports:                    opts.Ports,
				Selector:                 opts.Selector,
				ClusterIP:                opts.ClusterIP,
				PublishNotReadyAddresses: opts.PublishNotReadyAddress,
				LoadBalancerIP:           opts.LoadBalancerIP,
			},
		}
	}

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        opts.Name,
			Labels:      opts.Labels,
			Annotations: annotations,
		},
		Spec: corev1.ServiceSpec{
			Ports:                    opts.Ports,
			Selector:                 opts.Selector,
			Type:                     opts.Type,
			PublishNotReadyAddresses: opts.PublishNotReadyAddress,
			LoadBalancerIP:           opts.LoadBalancerIP,
		},
	}
}
