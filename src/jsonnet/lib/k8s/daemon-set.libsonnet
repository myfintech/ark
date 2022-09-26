function(
  name,
  podTemplate,
  restartPolicy='Always',
  volumes=[],
  annotations={},
) {
  local this = self,
  apiVersion: 'apps/v1',
  kind: 'DaemonSet',
  metadata: {
    name: name,
    labels: {
      app: name,
    },
  },
  spec: {
    selector: {
      matchLabels: this.metadata.labels,
    },
    updateStrategy: {
      rollingUpdate: {
        maxUnavailable: 1,
      },
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
        volumes: volumes,
      },
    },
  },
}
