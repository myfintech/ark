local serviceAccount = import 'lib/k8s/service-account.libsonnet';

function(
) {
  serverServiceAccount: serviceAccount(
    name='consul-server-svc-account',
    labels={ app: 'consul' },
  ),
}
