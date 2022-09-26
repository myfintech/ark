local configMap = import 'lib/k8s/config-map.libsonnet';

function() {
  serverConfigMap: configMap(
    name='consul-server-config',
    labels={ app: 'consul' },
    data={
      'central-config.json': std.toString({
        enable_central_service_config: true,
      }),
      'extra-from-values.json': std.toString({}),
      'proxy-defaults-config.json': std.toString({
        config_entries: {
          bootstrap: [
            {
              kind: 'proxy-defaults',
              name: 'global',
              config:
                {
                  envoy_dogstatsd_url: 'udp://127.0.0.1:9125',
                },

            },
          ],
        },
      }),
    },
  ),
}
