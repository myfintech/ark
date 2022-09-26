function(
  name,
  replicas,
  podTemplate,
  annotations={},
  restartPolicy='Always',
) {
  local this = self,
  apiVersion: 'apps/v1',
  kind: 'StatefulSet',
  metadata: {
    name: name,
    labels: {
      app: name,
    },
    annotations: annotations,
  },
  spec: {
    replicas: replicas,
    selector: {
      matchLabels: this.metadata.labels,
    },
    template: podTemplate {
      metadata+: {
        name: this.metadata.name,
        labels+: this.metadata.labels,
        annotations: annotations,
      },
    },
  },
}
