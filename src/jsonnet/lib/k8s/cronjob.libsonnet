function(
  name,
  schedule,
  jobTemplate,
  restartPolicy='OnFailure',
  concurrencyPolicy='Allow'
) {
  metadata: {
    name: name,
  },
  apiVersion: 'batch/v1beta1',
  kind: 'CronJob',
  spec: {
    schedule: schedule,
    jobTemplate: job {
      spec+: {
        backoffLimit: 6,
        template+: {
          metadata+: {
            name: self.metadata.name,
          },
          spec+: {
            restartPolicy: restartPolicy,
          },
        },
      },
    },
    concurrencyPolicy: concurrencyPolicy,
  },
}
