function(
  name,
  hosts=[],
  host,
  port,
) {
  apiVersion: 'networking.istio.io/v1alpha3',
  kind: 'VirtualService',
  metadata: {
    name: name,
  },
  spec: {
    hosts: hosts,
    http: [{
      route: [{
        destination: {
          host: host,
          port: {
            number: port,
          },
        },
      }],
    }],
  },
}
