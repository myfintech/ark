function(
  name,
  securityPolicy
) {
  apiVersion: 'cloud.google.com/v1beta1',
  kind: 'BackendConfig',
  metadata: {
    name: name,
  },
  spec: {
    securityPolicy: securityPolicy,
    cdn: {
      enabled: false,
      includeHost: true,
      includeProtocol: true,
      includeQueryString: false,
    },
  },
}
