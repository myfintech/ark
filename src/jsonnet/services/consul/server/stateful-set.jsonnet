local container = import 'lib/k8s/container.libsonnet';
local podTemplate = import 'lib/k8s/pod-template.libsonnet';
local statefulSet = import 'lib/k8s/stateful-set.libsonnet';

function(
  serviceAccountName,
  serverConfigMap,
) statefulSet(
  name='consul-server',
  replicas=1,
  podTemplate={},
) {
  local statefulSetContainer = container(
    name='consul-server',
    image='consul:1.7.2',
    command=[
      '/bin/sh',
      '-ec',
      'CONSUL_FULLNAME="consul"\n\nexec /bin/consul agent \\\n  -advertise="${POD_IP}" \\\n  -bind=0.0.0.0 \\\n  -bootstrap-expect=1 \\\n  -client=0.0.0.0 \\\n  -config-dir=/consul/config \\\n  -datacenter=dc1 \\\n  -data-dir=/consul/data \\\n  -domain=consul \\\n  -hcl="connect { enabled = true }" \\\n  -ui \\\n  -retry-join=${CONSUL_FULLNAME}-server-0.${CONSUL_FULLNAME}-server.${NAMESPACE}.svc \\\n  -server\n',
    ],
    env=[
      {
        name: 'POD_IP',
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
    ],
    ports=[
      {
        containerPort: 8500,
        name: 'http',
      },
      {
        containerPort: 8301,
        name: 'serflan',
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
    lifecycle: {
      preStop: {
        exec: {
          command: [
            '/bin/sh',
            '-c',
            'consul leave',
          ],
        },
      },
    },
    readinessProbe: {
      failureThreshold: 2,
      initialDelaySeconds: 5,
      periodSeconds: 3,
      successThreshold: 1,
      timeoutSeconds: 5,
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

  local statefulSetPodTemplate = podTemplate(
    containers=[statefulSetContainer],
    serviceAccountName=serviceAccountName,
  ) {
    metadata+: {
      labels+: {
        component: 'server',
        app: 'consul',
        hasDNS: 'true',
      },
      annotations+: {
        'consul.hashicorp.com/connect-inject': 'false',
      },
    },
    spec+: {
      affinity: {
        podAntiAffinity: {
          requiredDuringSchedulingIgnoredDuringExecution: [
            {
              labelSelector: {
                matchLabels: {
                  app: 'consul',
                  component: 'server',
                },
              },
              topologyKey: 'kubernetes.io/hostname',
            },
          ],
        },
      },
      securityContext: {
        fsGroup: 1000,
      },
      volumes+: [
        {
          configMap: {
            name: serverConfigMap,
          },
          name: 'config',
        },
      ],
    },
  },
  spec+: {
    podManagementPolicy: 'Parallel',
    volumeClaimTemplates: [
      {
        metadata: {
          name: 'data',
        },
        spec: {
          accessModes: [
            'ReadWriteOnce',
          ],
          resources: {
            requests: {
              storage: '10Gi',
            },
          },
        },
      },
    ],
    selector+: {
      matchLabels+: {
        app: 'consul',
        hasDNS: 'true',
        component: 'server',
      },
    },
    serviceName: 'consul-server',
    template+: statefulSetPodTemplate,
  },
  metadata+: {
    labels+: {
      component: 'server',
      app: 'consul',
    },
  },
}
