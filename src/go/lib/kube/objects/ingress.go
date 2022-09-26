package objects

import (
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// prefix = v1.PathType("Prefix")
	// exact                  = v1.PathType("Exact")
	implementationSpecific = v1.PathType("ImplementationSpecific")
)

// IngressOptions represents fields that can be passed in to create an ingress
type IngressOptions struct {
	Name               string            `json:"name"`
	Rules              []v1.IngressRule  `json:"rules"`
	TLS                []v1.IngressTLS   `json:"tls,omitempty"`
	Annotations        map[string]string `json:"annotations,omitempty"`
	GoogleGlobalIPName string            `json:"googleGlobalIPName,omitempty"`
}

// Ingress returns a pointer to an ingress object
func Ingress(opts IngressOptions) *v1.Ingress {
	// Need to define annotations because it could be nil
	annotations := make(map[string]string, 0)
	if opts.GoogleGlobalIPName != "" {
		annotations["kubernetes.io/ingress.global-static-ip-name"] = opts.GoogleGlobalIPName
	}
	for k, v := range opts.Annotations {
		annotations[k] = v
	}

	return &v1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        opts.Name,
			Annotations: annotations,
		},
		Spec: v1.IngressSpec{
			TLS:   opts.TLS,
			Rules: opts.Rules,
		},
	}
}

// BuildIngressRule returns a constructed ingress rule from a host, service name, and service port
func BuildIngressRule(host, serviceName, path string, servicePort int32) v1.IngressRule {
	pathType := implementationSpecific
	return v1.IngressRule{
		Host: host,
		IngressRuleValue: v1.IngressRuleValue{
			HTTP: &v1.HTTPIngressRuleValue{
				Paths: []v1.HTTPIngressPath{
					{
						Path:     path,
						PathType: &pathType,
						Backend: v1.IngressBackend{
							Service: &v1.IngressServiceBackend{
								Name: serviceName,
								Port: v1.ServiceBackendPort{
									Number: servicePort,
								},
							},
						},
					},
				},
			},
		},
	}
}
