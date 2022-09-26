function(
  name,
  kind,
  rules,
  labels={},
) {
  metadata: {
    name: name,
    labels: labels,
  },
  apiVersion: 'rbac.authorization.k8s.io/v1',
  kind: kind,
  rules: rules,
}
