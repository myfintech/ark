local service = import 'lib/k8s/service.libsonnet';

function() {
  consulDNSService: service(
    name='consul-dns',
    ports=[
      {
        name: 'dns-tcp',
        port: 53,
        protocol: 'TCP',
        targetPort: 'dns-tcp',
      },
      {
        name: 'dns-udp',
        port: 53,
        protocol: 'UDP',
        targetPort: 'dns-udp',
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
        app: 'consul',
        hasDNS: 'true',
      },
    },
  },
}
