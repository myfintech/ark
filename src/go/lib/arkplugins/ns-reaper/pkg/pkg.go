package pkg

import (
	"bytes"

	v1 "k8s.io/api/networking/v1"

	"github.com/myfintech/ark/src/go/lib/kube/microservice"
	"github.com/myfintech/ark/src/go/lib/kube/objects"
	rbacv1 "k8s.io/api/rbac/v1"
)

func NewNsReaperManifest(opts microservice.Options) (string, error) {
	manifest := new(objects.Manifest)
	buff := new(bytes.Buffer)

	msObjects := microservice.NewMicroService(opts)

	for _, object := range msObjects.CoreList.Items {
		manifest.CoreList.Items = append(manifest.CoreList.Items, object)
	}

	ingressRule := objects.BuildIngressRule("reaper.eph.mantl.dev", opts.Name, "", opts.ServicePort)

	ingressOptions := objects.IngressOptions{
		Name:        opts.Name,
		Rules:       []v1.IngressRule{ingressRule},
		Annotations: map[string]string{"kubernetes.io/ingress.class": "nginx"},
	}

	ingress := objects.Ingress(ingressOptions)

	manifest.Append(ingress)

	clusterRole := objects.ClusterRole(objects.RoleOptions{
		Name: opts.Name,
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{
					"namespaces",
				},
				Verbs: []string{"list", "delete"},
			},
		},
	})

	clusterRoleBinding := objects.ClusterRoleBinding(objects.RoleBindingOptions{
		Name: opts.Name,
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      opts.Name,
				Namespace: "sre",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     opts.Name,
		},
	})

	manifest.Append(
		clusterRole,
		clusterRoleBinding,
	)

	if err := manifest.Serialize(buff); err != nil {
		return "", err
	}
	return buff.String(), nil
}
