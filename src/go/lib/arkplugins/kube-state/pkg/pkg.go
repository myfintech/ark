package pkg

import (
	"bytes"

	"github.com/myfintech/ark/src/go/lib/kube/objects"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type KubeStateOptions struct {
	Name     string `json:"name"`
	Image    string `json:"image"`
	Replicas int32  `json:"replicas"`
}

func NewKubeStateManifest(opts KubeStateOptions) (string, error) {
	manifest := new(objects.Manifest)
	buff := new(bytes.Buffer)

	serviceAccount := objects.ServiceAccount(objects.ServiceAccountOptions{
		Name: opts.Name,
	})

	clusterRole := buildClusterRole(&opts)

	clusterRoleBinding := objects.ClusterRoleBinding(objects.RoleBindingOptions{
		Name: opts.Name,
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      opts.Name,
				Namespace: opts.Name,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     opts.Name,
		},
	})

	container := objects.Container(objects.ContainerOptions{
		Name:  opts.Name,
		Image: opts.Image,
		Ports: []corev1.ContainerPort{
			{
				Name:          "http-metrics",
				ContainerPort: 8080,
			},
			{
				Name:          "telemetry",
				ContainerPort: 8081,
			},
		},
		ImagePullPolicy: corev1.PullIfNotPresent,
		LivenessProbe: corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/healthz",
					Port: intstr.FromInt(8080),
				},
			},
			InitialDelaySeconds: 5,
			TimeoutSeconds:      5,
		},
		ReadinessProbe: corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/healthz",
					Port: intstr.FromInt(8080),
				},
			},
			InitialDelaySeconds: 5,
			TimeoutSeconds:      5,
		},
	})

	podTemplate := objects.PodTemplate(objects.PodTemplateOptions{
		Name:               opts.Name,
		Containers:         []corev1.Container{*container},
		RestartPolicy:      "Always",
		ServiceAccountName: opts.Name,
		Labels: map[string]string{
			"app": opts.Name,
		},
	})

	deployment := objects.Deployment(objects.DeploymentOptions{
		Name:        opts.Name,
		Replicas:    &opts.Replicas,
		PodTemplate: *podTemplate,
	})

	service := objects.Service(objects.ServiceOptions{
		Name: opts.Name,
		Ports: []corev1.ServicePort{
			{
				Name:       "http-metrics",
				Protocol:   "TCP",
				Port:       8080,
				TargetPort: intstr.FromString("http-metrics"),
			},
			{
				Name:       "telemetry",
				Protocol:   "TCP",
				Port:       8081,
				TargetPort: intstr.FromString("telemetry"),
			},
		},
		Type: "ClusterIP",
		Selector: map[string]string{
			"app": opts.Name,
		},
	})

	manifest.Append(
		serviceAccount,
		clusterRole,
		clusterRoleBinding,
		deployment,
		service,
	)

	if err := manifest.Serialize(buff); err != nil {
		return "", err
	}
	return buff.String(), nil
}
