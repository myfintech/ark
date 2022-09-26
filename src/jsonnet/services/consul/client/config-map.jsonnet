local configMap = import 'lib/k8s/config-map.libsonnet';

function() {
  clientConfigMap: configMap(
    name='consul-client-config',
    labels={ app: 'consul' },
    data={
      'central-config.json': std.toString({
        enable_central_service_config: true,
      }),
      'extra-from-values.json': std.toString({}),
    },
  ),
}
