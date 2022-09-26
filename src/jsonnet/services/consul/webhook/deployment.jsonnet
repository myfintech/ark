local container = import 'lib/k8s/container.libsonnet';
local deployment = import 'lib/k8s/deployment.libsonnet';
local podTemplate = import 'lib/k8s/pod-template.libsonnet';

function(
  serviceAccountName,
) deployment(
  name='consul-connect-injector-webhook-deployment',
  replicas=1,
  podTemplate={},
) {
  local deploymentContainer = container(
    name='sidecar-injector',
    image='hashicorp/consul-k8s:0.14.0',
    command=[
      '/bin/sh',
      '-ec',
      'CONSUL_FULLNAME="consul"\n\nconsul-k8s inject-connect \\\n  -default-inject=true \\\n  -consul-image="consul:1.7.2" \\\n  -consul-k8s-image="hashicorp/consul-k8s:0.14.0" \\\n  -listen=:8080 \\\n  -enable-central-config=true \\\n  -default-protocol="http" \\\n  -allow-k8s-namespace="*" \\\n  -tls-auto=${CONSUL_FULLNAME}-connect-injector-cfg \\\n  -tls-auto-hosts=${CONSUL_FULLNAME}-connect-injector-svc,${CONSUL_FULLNAME}-connect-injector-svc.${NAMESPACE},${CONSUL_FULLNAME}-connect-injector-svc.${NAMESPACE}.svc\n',
    ],
    env=[
      {
        name: 'NAMESPACE',
        valueFrom: {
          fieldRef: {
            fieldPath: 'metadata.namespace',
          },
        },
      },
    ],
  ) {
    livenessProbe: {
      failureThreshold: 2,
      httpGet: {
        path: '/health/ready',
        port: 8080,
        scheme: 'HTTPS',
      },
      initialDelaySeconds: 1,
      periodSeconds: 2,
      successThreshold: 1,
      timeoutSeconds: 5,
    },
    readinessProbe: {
      failureThreshold: 2,
      httpGet: {
        path: '/health/ready',
        port: 8080,
        scheme: 'HTTPS',
      },
      initialDelaySeconds: 2,
      periodSeconds: 2,
      successThreshold: 1,
      timeoutSeconds: 5,
    },
  },

  local deploymentPodTemplate = podTemplate(
    containers=[deploymentContainer],
    serviceAccountName=serviceAccountName,
  ) {
    metadata+: {
      labels+: {
        component: 'connect-injector',
        app: 'consul',
      },
      annotations+: {
        'consul.hashicorp.com/connect-inject': 'false',
      },
    },
  },
  metadata+: {
    labels+: {
      app: 'consul',
    },
  },
  spec+: {
    selector+: {
      matchLabels+: {
        component: 'connect-injector',
        app: 'consul',
      },
    },
    template+: deploymentPodTemplate,
  },
}
