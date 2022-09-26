function(
  name,
  replicas,
  podTemplate,
  annotations={},
  restartPolicy='Always',
) {
  local this = self,
  apiVersion: 'apps/v1',
  kind: 'Deployment',
  metadata: {
    labels: {
      app: name,
    },
    annotations: annotations,
    name: name,
  },
  spec: {
    replicas: replicas,
    selector: {
      matchLabels: {
        app: name,
      },
    },
    strategy: {
      type: 'RollingUpdate',
    },
    template: podTemplate {
      metadata+: {
        name: this.metadata.name,
        labels+: this.metadata.labels,
        annotations: annotations,
      },
      spec+: {
        restartPolicy: restartPolicy,
      },
    },
  },
}
