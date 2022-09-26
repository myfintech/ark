package objects

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RoleOptions represents fields that can be passed in to create a role or cluster role
type RoleOptions struct {
	Name   string
	Rules  []rbacv1.PolicyRule
	Labels map[string]string
}

// Role returns a pointer to a role
func Role(opts RoleOptions) *rbacv1.Role {
	return &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   opts.Name,
			Labels: opts.Labels,
		},
		Rules: opts.Rules,
	}
}

// ClusterRole returns a pointer to a cluster role
func ClusterRole(opts RoleOptions) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   opts.Name,
			Labels: opts.Labels,
		},
		Rules: opts.Rules,
	}
}
