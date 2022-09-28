package objects

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RoleBindingOptions struct {
	Name     string
	Subjects []rbacv1.Subject
	RoleRef  rbacv1.RoleRef
	Labels   map[string]string
}

// RoleBinding returns a pointer to a role binding
func RoleBinding(opts RoleBindingOptions) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   opts.Name,
			Labels: opts.Labels,
		},
		Subjects: opts.Subjects,
		RoleRef:  opts.RoleRef,
	}
}

// ClusterRoleBinding returns a pointer to a cluster role binding
func ClusterRoleBinding(opts RoleBindingOptions) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   opts.Name,
			Labels: opts.Labels,
		},
		Subjects: opts.Subjects,
		RoleRef:  opts.RoleRef,
	}
}
