function(
  name,
  labels={}
) {
  apiVersion: 'v1',
  kind: 'ServiceAccount',
  metadata: {
    name: name,
    labels: labels,
  },
}
