local serviceAccount = import 'lib/k8s/service-account.libsonnet';

function(
) {
  clientServiceAccount: serviceAccount(
    name='consul-client-svc-account',
    labels={ app: 'consul' },
  ),
}
