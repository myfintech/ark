function(
  name,
  data,
  labels={},
) {
  apiVersion: 'v1',
  kind: 'ConfigMap',
  metadata: {
    name: name,
    labels: labels,
  },
  data: data,
}
