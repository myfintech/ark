function(
  name,
  ports,
  backendConfig='',
  type='ClusterIP',
  selector={},
  labels={},
  annotations={},
) {
  apiVersion: 'v1',
  kind: 'Service',
  metadata: {
    name: name,
    labels: labels {
      app: name,
    },
    annotations: annotations + if backendConfig != '' then {
      'beta.cloud.google.com/backend-config': backendConfig,
    } else {},
  },
  spec: {
    ports: ports,
    type: type,
    selector: selector {
      app: name,
    },
  },
}
