function(
  name,
  rules,
  tls=[],
  annotations={},
  googleGlobalIPName=''
) {
  apiVersion: 'extensions/v1beta1',
  kind: 'Ingress',
  metadata: {
    name: name,
    annotations: annotations + if googleGlobalIPName != '' then {
      'kubernetes.io/ingress.global-static-ip-name': googleGlobalIPName,
    } else {},
  },
  spec: {
    tls: tls,
    rules: rules,
  },
}
