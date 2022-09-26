local container = import '../k8s/container.libsonnet';
local deployment = import '../k8s/deployment.libsonnet';
local ingress = import '../k8s/ingress.libsonnet';
local podTemplate = import '../k8s/pod-template.libsonnet';
local service = import '../k8s/service.libsonnet';

local ingressRule(host, serviceName, servicePort) = {
  host: host,
  http: {
    paths: [
      {
        backend: {
          serviceName: serviceName,
          servicePort: servicePort,
        },
      },
    ],
  },
};

local liveSyncLabels(liveSyncEnabled, targetAddress) = {
  'ark.live.sync.enabled': std.toString(liveSyncEnabled),
  'ark.target.address': targetAddress,
};

local applyArkLabels(deployment, liveSyncEnabled, targetAddress) = deployment {
  metadata+: {
    labels+: liveSyncLabels(liveSyncEnabled, targetAddress),
  },
  spec+: {
    template+: {
      metadata+: {
        labels+: liveSyncLabels(liveSyncEnabled, targetAddress),
      },
    },
  },
};

function(
  name,
  image,
  command,
  replicas=1,
  env=[],
  port=3000,
  servicePort=80,
  targetAddress='',
  hostNetwork=false,
  serviceType='ClusterIP',
  liveSyncEnabled=false,
  liveSyncRestartMode='auto',
  hosts=[],
) {
  deployment: applyArkLabels(
    deployment(
      name=name,
      replicas=replicas,
      podTemplate=podTemplate(
        hostNetwork=hostNetwork,
        containers=[container(
          name=name,
          image=image,
          command=(if liveSyncEnabled then ['/usr/local/bin/ark-entrypoint'] else command),
          args=(if liveSyncEnabled then command else []),
        ) {
          env+: env + [
            {
              name: 'PORT',
              value: std.toString(port),
            },
            {
              name: 'ARK_EP_RESTART_MODE',
              value: liveSyncRestartMode,
            },
          ],
        }],
      )
    ), liveSyncEnabled, targetAddress,
  ),
  service: service(
    name=name,
    type=serviceType,
    ports=[{
      name: 'http',
      port: servicePort,
      targetPort: port,
    }],
    labels=liveSyncLabels(liveSyncEnabled, targetAddress),
  ),
  ingress: ingress(
    name=name,
    rules=[ingressRule(host, name, 'http') for host in hosts],
    annotations={
      'kubernetes.io/ingress.class': 'nginx',
    },
  ),
}
