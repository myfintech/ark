local webhook = import 'lib/k8s/mutating-webhook.libsonnet';

function(

) {
  mutatingWebhook: webhook(
    name='consul-connect-injector-cfg',
    labels={ app: 'consul' },
  ) {
    webhooks: [
      {
        clientConfig: {
          caBundle: '',
          service: {
            name: 'consul-connect-injector-svc',
            namespace: 'consul',
            path: '/mutate',
          },
        },
        name: 'consul-connect-injector.consul.hashicorp.com',
        rules: [
          {
            apiGroups: [
              '',
            ],
            apiVersions: [
              'v1',
            ],
            operations: [
              'CREATE',
            ],
            resources: [
              'pods',
            ],
          },
        ],
      },
    ],
  },
}
