package pkg

import (
	"bytes"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/myfintech/ark/src/go/lib/kube/mutations"
	"github.com/myfintech/ark/src/go/lib/kube/objects"
)

// Options is the available input options for deploying VDS
type Options struct {
	mutations.VaultConfig

	Name     string `json:"name"`
	Address  string `json:"address"`
	Image    string `json:"image"`
	Port     int    `json:"port"`
	Replicas int32  `json:"replicas"`
}

// NewManifest produces a Kubernetes compatible resource deployment manifest
func NewManifest(opts Options) (string, error) {
	name := opts.Name
	manifest := new(objects.Manifest)
	buff := new(bytes.Buffer)
	defaultLabels := map[string]string{"app": name}

	var serviceAccountName string
	var shortEnv string
	switch opts.Environment {
	case "integration":
		serviceAccountName = fmt.Sprintf("%s-int", name)
		shortEnv = "int"
	case "qa":
		serviceAccountName = name
		shortEnv = opts.Environment
	case "uat":
		serviceAccountName = fmt.Sprintf("%s-%s", name, opts.Environment)
		shortEnv = opts.Environment
	case "production":
		serviceAccountName = fmt.Sprintf("%s-prod", name)
		shortEnv = "prod"
	case "demo":
		serviceAccountName = name
		shortEnv = opts.Environment
	}

	opts.Role = serviceAccountName

	// service account used for Vault authentication and access to ingress API
	serviceAccount := objects.ServiceAccount(objects.ServiceAccountOptions{
		Name:   serviceAccountName,
		Labels: defaultLabels,
	})

	// cluster role for access to ingress API
	clusterRole := objects.ClusterRole(objects.RoleOptions{
		Name: name,
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"networking.k8s.io",
				},
				Resources: []string{
					"ingresses",
				},
				Verbs: []string{
					"get",
					"list",
					"watch",
					"create",
					"update",
					"patch",
					"delete",
				},
			},
		},
	})

	// cluster role binding for cluster role and service account
	clusterRoleBinding := objects.ClusterRoleBinding(objects.RoleBindingOptions{
		Name: name,
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName,
				Namespace: "sre",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     name,
		},
	})

	// figure out port and listening address and create env var
	var probePort intstr.IntOrString
	if opts.Port != 0 {
		probePort = intstr.FromInt(opts.Port)
	} else {
		probePort = intstr.FromInt(3000)
	}

	var listeningAddress string
	if opts.Address != "" {
		listeningAddress = fmt.Sprintf("%s:%d", opts.Address, probePort.IntVal)
	} else {
		listeningAddress = fmt.Sprintf("0.0.0.0:%d", probePort.IntVal)
	}

	env := []corev1.EnvVar{
		{
			Name:  "VDS_ADDRESS",
			Value: listeningAddress,
		},
	}

	// container for service
	container := objects.Container(objects.ContainerOptions{
		Name:    name,
		Image:   opts.Image,
		Command: []string{"/usr/local/bin/vds"},
		Env:     env,
		LivenessProbe: corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/health/live",
					Port: probePort,
				},
			},
			InitialDelaySeconds: 5,
			TimeoutSeconds:      5,
		},
		ReadinessProbe: corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/health/ready",
					Port: probePort,
				},
			},
			InitialDelaySeconds: 5,
			TimeoutSeconds:      5,
		},
	})

	// template to wrap container
	podTemplate := objects.PodTemplate(objects.PodTemplateOptions{
		Name:               name,
		Containers:         []corev1.Container{*container},
		RestartPolicy:      "Always",
		ServiceAccountName: serviceAccountName,
		Labels:             defaultLabels,
	})

	// deployment with Vault init containers
	deployment := mutations.ApplyVault(
		true,
		opts.VaultConfig,
		objects.Deployment(objects.DeploymentOptions{
			Name:        name,
			Replicas:    &opts.Replicas,
			PodTemplate: *podTemplate,
		}),
	)

	// service resource
	service := objects.Service(objects.ServiceOptions{
		Name: name,
		Ports: []corev1.ServicePort{
			{
				Name:       "http",
				Port:       int32(80),
				TargetPort: probePort,
			},
		},
		Type:     "ClusterIP",
		Selector: defaultLabels,
	})

	// ingress for com and internal hostnames
	publicHostname := fmt.Sprintf("vds.%s.mantl.com", shortEnv)
	privateHostname := fmt.Sprintf("vds.%s.mantl.internal", shortEnv)
	hostnames := []string{publicHostname, privateHostname}
	rules := make([]networkingv1.IngressRule, 0)
	for _, host := range hostnames {
		rules = append(rules, objects.BuildIngressRule(
			host,
			name,
			"/",
			int32(80),
		))
	}

	ingress := objects.Ingress(objects.IngressOptions{
		Name:  name,
		Rules: rules,
		Annotations: map[string]string{
			"kubernetes.io/ingress.class": "nginx",
		},
	})

	// complete deployment manifest
	manifest.Append(
		serviceAccount,
		clusterRole,
		clusterRoleBinding,
		deployment,
		service,
		ingress,
	)

	if err := manifest.Serialize(buff); err != nil {
		return "", err
	}

	return buff.String(), nil
}
