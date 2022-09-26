function(
  name,
  activeDeadlineSeconds,
  hostNetwork,
  podTemplatem,
  annotations={},
  backoffLimit=0,
  restartPolicy='Never',
) {
  metadata: {
    name: name,
    annotations: annotations,
  },
  apiVersion: 'batch/v1',
  kind: 'Job',
  spec: {
    backoffLimit: backoffLimit,
    activeDeadlineSeconds: activeDeadlineSeconds,
    template: podTemplate {
      metadata+: {
        name: self.metadata.name,
        annotations+: self.metadata.annotations,
      },
      spec+: {
        restartPolicy: restartPolicy,
      },
    },
  },
}
