local configMaps = import './config-maps.jsonnet';
local container = import 'lib/k8s/container.libsonnet';
local daemonSet = import 'lib/k8s/daemon-set.libsonnet';
local podTemplate = import 'lib/k8s/pod-template.libsonnet';

function(
  serviceAccountName,
  clientConfigMap,
) daemonSet(
  name='consul',
  podTemplate={},
) {

  local daemonSetContainer = container(
    name='consul',
    image='consul:1.7.2',
    command=[
      '/bin/sh',
      '-ec',
      'CONSUL_FULLNAME="consul"\n\nexec /bin/consul agent \\\n  -node="${NODE}" \\\n  -advertise="${ADVERTISE_IP}" \\\n  -bind=0.0.0.0 \\\n  -client=0.0.0.0 \\\n  -node-meta=pod-name:${HOSTNAME} \\\n  -hcl="leave_on_terminate = true" \\\n  -hcl="ports { grpc = 8502 }" \\\n  -config-dir=/consul/config \\\n  -datacenter=dc1 \\\n  -data-dir=/consul/data \\\n  -retry-join="${CONSUL_FULLNAME}-server-0.${CONSUL_FULLNAME}-server.${NAMESPACE}.svc" \\\n  -domain=consul\n',
    ],
    env=[
      {
        name: 'ADVERTISE_IP',
        valueFrom: {
          fieldRef: {
            fieldPath: 'status.podIP',
          },
        },
      },
      {
        name: 'NAMESPACE',
        valueFrom: {
          fieldRef: {
            fieldPath: 'metadata.namespace',
          },
        },
      },
      {
        name: 'NODE',
        valueFrom: {
          fieldRef: {
            fieldPath: 'spec.nodeName',
          },
        },
      },
      {
        name: 'HOST_IP',
        valueFrom: {
          fieldRef: {
            fieldPath: 'status.hostIP',
          },
        },
      },
    ],
    ports=[
      {
        containerPort: 8500,
        hostPort: 8500,
        name: 'http',
      },
      {
        containerPort: 8502,
        hostPort: 8502,
        name: 'grpc',
      },
      {
        containerPort: 8301,
        name: 'serflan-tcp',
        protocol: 'TCP',
      },
      {
        containerPort: 8301,
        name: 'serflan-udp',
        protocol: 'UDP',
      },
      {
        containerPort: 8302,
        name: 'serfwan',
      },
      {
        containerPort: 8300,
        name: 'server',
      },
      {
        containerPort: 8600,
        name: 'dns-tcp',
        protocol: 'TCP',
      },
      {
        containerPort: 8600,
        name: 'dns-udp',
        protocol: 'UDP',
      },
    ],
  ) {
    readinessProbe: {
      exec: {
        command: [
          '/bin/sh',
          '-ec',
          'curl http://127.0.0.1:8500/v1/status/leader \\\n2>/dev/null | grep -E "".+""\n',
        ],
      },
    },
    volumeMounts: [
      {
        mountPath: '/consul/data',
        name: 'data',
      },
      {
        mountPath: '/consul/config',
        name: 'config',
      },
    ],
  },

  local daemonSetPodTemplate = podTemplate(
    containers=[daemonSetContainer],
    serviceAccountName=serviceAccountName,
  ) {
    metadata+: {
      annotations+: { 'consul.hashicorp.com/connect-inject': 'false' },
      labels+: {
        component: 'client',
        app: 'consul',
        hasDNS: 'true',
      },
    },
    spec+: {
      terminationGracePeriodSeconds: 10,
      volumes+: [
        {
          emptyDir: {},
          name: 'data',
        },
        {
          configMap: {
            name: clientConfigMap,
          },
          name: 'config',
        },
      ],
    },
  },
  spec+: {
    selector+: {
      matchLabels+: {
        hasDNS: 'true',
        component: 'client',
      },
    },
    template+: daemonSetPodTemplate,
  },
}
