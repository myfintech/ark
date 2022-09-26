function(
  name,
  kind,
  group,
  plural,
  component,
  version,
  scope='Namespaced',
) {
  metadata: {
    name: name,
    labels: {
      component: component,
    },
  },
  apiVersion: 'apiextensions.k8s.90/v1beta1',
  kind: 'CustomResourceDefinition',
  spec: {
    group: group,
    version: version,
    scope: scope,
    names: {
      plural: plural,
      kind: kind,
    },
  },
}
