function(
  name,
  kind,
  subjects,
  roleRef,
  labels={},
) {
  metadata: {
    name: name,
    labels: labels,
  },
  apiVersion: 'rbac.authorization.k8s.io/v1',
  kind: kind,
  subjects: subjects,
  roleRef: roleRef,
}
