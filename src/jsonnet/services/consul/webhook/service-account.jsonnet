local serviceAccount = import 'lib/k8s/service-account.libsonnet';

function(
) {
  webhookServiceAccount: serviceAccount(
    name='consul-connect-injector-webhook-svc-account',
    labels={ app: 'consul' },
  ),
}
