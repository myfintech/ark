local service = import 'lib/k8s/service.libsonnet';

function() {
  serverService: service(
    name='consul-server',
    annotations={ 'service.alpha.kubernetes.io/tolerate-unready-endpoints': 'true' },
    ports=[
      {
        name: 'http',
        port: 8500,
        targetPort: 8500,
      },
      {
        name: 'serflan-tcp',
        port: 8301,
        protocol: 'TCP',
        targetPort: 8301,
      },
      {
        name: 'serflan-udp',
        port: 8301,
        protocol: 'UDP',
        targetPort: 8301,
      },
      {
        name: 'serfwan-tcp',
        port: 8302,
        protocol: 'TCP',
        targetPort: 8302,
      },
      {
        name: 'serfwan-udp',
        port: 8302,
        protocol: 'UDP',
        targetPort: 8301,
      },
      {
        name: 'server',
        port: 8300,
        targetPort: 8300,
      },
      {
        name: 'dns-tcp',
        port: 8600,
        protocol: 'TCP',
        targetPort: 'dns-tcp',
      },
      {
        name: 'dns-udp',
        port: 8600,
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
        component: 'server',
        app: 'consul',
      },
      clusterIP: 'None',
      publishNotReadyAddresses: true,
    },
  },

  uiService: service(
    name='consul-ui',
    ports=[
      {
        name: 'http',
        port: 80,
        targetPort: 8500,
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
        component: 'server',
        app: 'consul',
      },
    },
  },
}
