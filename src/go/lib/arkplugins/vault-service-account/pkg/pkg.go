package pkg

import (
	"bytes"

	"github.com/myfintech/ark/src/go/lib/kube/objects"
	rbacv1 "k8s.io/api/rbac/v1"
)

func NewVaultServiceAccountManifest() (string, error) {
	manifest := objects.Manifest{}
	buff := new(bytes.Buffer)

	saOpts := objects.ServiceAccountOptions{
		Name: "vault-token-reviewer",
	}
	sa := objects.ServiceAccount(saOpts)

	crbOpts := objects.RoleBindingOptions{
		Name: "vault-token-reviewer",
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      saOpts.Name,
				Namespace: "sre", // make sure the service account actually gets deployed here
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "system:auth-delegator",
		},
	}
	crb := objects.ClusterRoleBinding(crbOpts)

	manifest.Append(sa, crb)

	if err := manifest.Serialize(buff); err != nil {
		return "", err
	}
	return buff.String(), nil
}
