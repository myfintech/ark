function(
  name,
  host,
  version,
) {
  apiVersion: 'networking.istio.io/v1alpha3',
  kind: 'DestinationRule',
  metadata: {
    name: name,
  },
  spec: {
    host: host,
    subsets: [{
      name: {
        labels: {
          version: version,
        },
      },
    }],
  },
}
