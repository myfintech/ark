function(
  name,
  labels={},
) {
  metadata: {
    name: name,
    labels: labels,
  },
  apiVersion: 'v1',
  kind: 'Namespace',
}
