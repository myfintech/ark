local service = import 'lib/k8s/service.libsonnet';

function() {
  webhookService: service(
    name='consul-connect-injector-svc',
    ports=[
      {
        port: 443,
        targetPort: 8080,
      },
    ],

  ) {
    metadata+: {
      labels+: {
        app: 'consul',
      },
    },
    spec+: {
      selector+: {
        component: 'connect-injector',
        app: 'consul',
      },
    },
  },
}
