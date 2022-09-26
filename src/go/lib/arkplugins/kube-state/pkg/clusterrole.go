package pkg

import (
	"github.com/myfintech/ark/src/go/lib/kube/objects"
	rbacv1 "k8s.io/api/rbac/v1"
)

func buildClusterRole(opts *KubeStateOptions) *rbacv1.ClusterRole {
	commonVerbs := []string{"list", "watch"}

	return objects.ClusterRole(objects.RoleOptions{
		Name: opts.Name,
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{
					"configmaps",
					"secrets",
					"nodes",
					"pods",
					"services",
					"resourcequotas",
					"replicationcontrollers",
					"limitranges",
					"persistentvolumeclaims",
					"persistentvolumes",
					"namespaces",
					"endpoints",
				},
				Verbs: commonVerbs,
			},
			{
				APIGroups: []string{
					"extensions",
				},
				Resources: []string{
					"daemonsets",
					"deployments",
					"replicasets",
					"ingresses", // this is removed in newer versions as it's moved to the networking.k8s.io API
				},
				Verbs: []string{
					"list",
					"watch",
				},
			},
			{
				APIGroups: []string{
					"apps",
				},
				Resources: []string{
					"daemonsets",
					"deployments",
					"replicasets",
					"statefulsets",
				},
				Verbs: commonVerbs,
			},
			{
				APIGroups: []string{
					"batch",
				},
				Resources: []string{
					"cronjobs",
					"jobs",
				},
				Verbs: commonVerbs,
			},
			{
				APIGroups: []string{
					"autoscaling",
				},
				Resources: []string{
					"horizontalpodautoscalers",
				},
				Verbs: commonVerbs,
			},
			{
				APIGroups: []string{
					"authentication.k8s.io",
				},
				Resources: []string{
					"tokenreviews",
				},
				Verbs: []string{
					"create",
				},
			},
			{
				APIGroups: []string{
					"authorization.k8s.io",
				},
				Resources: []string{
					"subjectaccessreviews",
				},
				Verbs: []string{
					"create",
				},
			},
			{
				APIGroups: []string{
					"policy",
				},
				Resources: []string{
					"poddisruptionbudgets",
				},
				Verbs: commonVerbs,
			},
			{
				APIGroups: []string{
					"certificates.k8s.io",
				},
				Resources: []string{
					"certificatesigningrequests",
				},
				Verbs: commonVerbs,
			},
			{
				APIGroups: []string{
					"storage.k8s.io",
				},
				Resources: []string{
					"storageclasses",
					"volumeattachments",
				},
				Verbs: commonVerbs,
			},
			{
				APIGroups: []string{
					"admissionregistration.k8s.io",
				},
				Resources: []string{
					"mutatingwebhookconfigurations",
					"validatingwebhookconfigurations",
				},
				Verbs: commonVerbs,
			},
			{
				APIGroups: []string{
					"networking.k8s.io",
				},
				Resources: []string{
					"networkpolicies",
					"ingresses",
				},
				Verbs: commonVerbs,
			},
			{
				APIGroups: []string{
					"coordination.k8s.io",
				},
				Resources: []string{
					"leases",
				},
				Verbs: commonVerbs,
			},
		},
	})
}
