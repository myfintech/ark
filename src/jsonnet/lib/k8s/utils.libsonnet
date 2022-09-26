local applySingleHostAffinity(appName, deployment, disabled=false) = if disabled then deployment else deployment {
  spec+: {
    template+: {
      spec+: {
        affinity: {
          podAntiAffinity: {
            requiredDuringSchedulingIgnoredDuringExecution: [{
              labelSelector: {
                matchExpressions: [{
                  key: 'app',
                  operator: 'In',
                  values: [appName],
                }],
              },
              topologyKey: 'kubernetes.io/hostname',
            }],
          },
        },
      },
    },
  },
};

local applyGoogleCloudServiceAccount(deployment, filename='/etc/google/service_account.json', mountAsSecret=false, disabled=false) = if disabled then deployment else deployment {
  local applyUpdateToContainer(container) = container {
    env+: [{
      name: 'GOOGLE_APPLICATION_CREDENTIALS',
      value: filename,
    }],
    volumeMounts+: [{
      name: 'google-service-account',
      mountPath: '/etc/google',
    }],
  },
  spec+: {
    template+: {
      spec+: {
        containers: [applyUpdateToContainer(container) for container in deployment.spec.template.spec.containers],
        initContainers: [applyUpdateToContainer(container) for container in deployment.spec.template.spec.initContainers],
        volumes+: [{
          name: 'google-service-account',
        } + if mountAsSecret then { secret: {
          secretName: 'google-service-account',
        } } else { emptyDir: {} }],
      },
    },
  },
};

local vaultImage = 'gcr.io/managed-infrastructure/mantl/vault-auth:5610037';

local applyVaultInitContainer(deployment) = deployment {
  spec+: {
    template+: {
      spec+: {
        initContainers+: [{
          name: 'vault-login',
          image: vaultImage,
          command: [
            '/usr/local/bin/vault-auth',
            'login',
          ],
        }],
      },
    },
  },
};

local applyVaultAutoRenewContainer(deployment) = deployment {
  spec+: {
    template+: {
      spec+: {
        containers+: [{
          name: 'vault-auto-renew',
          image: vaultImage,
          command: [
            '/usr/local/bin/vault-auth',
            'renew',
          ],
          lifecycle: {
            preStop: {
              exec: {
                command: [
                  '/usr/local/bin/vault-auth',
                  'revoke',
                ],
              },
            },
          },
        }],
      },
    },
  },
};

local applyVaultConfigToContainer(vault, container) = container {
  env+: [
    {
      name: 'AUTH_SERVICE_ADDR',
      value: 'http://auth-service.vault:5000',
    },
    {
      name: 'VAULT_CONFIG',
      value: '/etc/vault/config.json',
    },
    {
      name: 'VAULT_ADDR',
      value: vault.address,
    },
    {
      name: 'VAULT_TEAM',
      value: vault.team,
    },
    {
      name: 'VAULT_ENV',
      value: vault.environment,
    },
    {
      name: 'VAULT_APP',
      value: vault.app,
    },
    {
      name: 'TOKEN_TTL_INCREMENT',
      value: '86400',
    },
    {
      name: 'VAULT_DEFAULT_CONFIG',
      value: vault.defaultConfig,
    },
    {
      name: 'POD_NAMESPACE',
      valueFrom: {
        fieldRef: {
          fieldPath: 'metadata.namespace',
        },
      },
    },
  ],
  volumeMounts+: [{
    name: 'vault-config',
    mountPath: '/etc/vault',
  }],
};

local applyVaultConfigToDeployment(vault, deployment) = deployment {
  local annotations = {
    'vault.app': vault.app,
    'vault.team': vault.team,
  },
  metadata+: {
    annotations+: annotations,
  },
  spec+: {
    template+: {
      metadata+: {
        annotations+: annotations,
      },
      spec+: {
        containers: [applyVaultConfigToContainer(vault, container) for container in deployment.spec.template.spec.containers],
        initContainers: [applyVaultConfigToContainer(vault, container) for container in deployment.spec.template.spec.initContainers],
        volumes+: [{
          name: 'vault-config',
          emptyDir: {},
        }],
      },
    },
  },
};

local applyVault(vault, deployment) = applyVaultConfigToDeployment(
  vault,
  applyVaultAutoRenewContainer(
    applyVaultInitContainer(
      deployment
    )
  )
);

local applyLivenessAndReadinessProbes(livePath, livePort, readyPath, readyPort, deployment) = deployment {
  local applyUpdateToContainer(container) = container {
    livenessProbe: {
      httpGet: {
        path: livePath,
        port: livePort,
      },
      initialDelaySeconds: 5,
      timeoutSeconds: 1,
      periodSeconds: 15,
    },
    readinessProbe: {
      httpGet: {
        path: readyPath,
        port: readyPort,
      },
      initialDelaySeconds: 5,
      timeoutSeconds: 1,
      periodSeconds: 15,
    },
  },
  spec+: {
    template+: {
      spec+: {
        containers: [applyUpdateToContainer(container) for container in deployment.spec.template.spec.containers],
      },
    },
  },
};

local applyImagePullPolicy(policy, deployment) = deployment {
  local applyUpdateToContainer(container) = container {
    imagePullPolicy: policy,
  },
  spec+: {
    template+: {
      spec+: {
        containers: [applyUpdateToContainer(container) for container in deployment.spec.template.spec.containers],
      },
    },
  },
};

local applyResourceConstraints(resourceConstraints) = function(deployment) deployment {
  local applyUpdateToContainer(container) = container {
    resources: resourceConstraints,
  },
  spec+: {
    template+: {
      spec+: {
        containers: [applyUpdateToContainer(container) for container in deployment.spec.template.spec.containers],
      },
    },
  },
};

{
  applyVault: applyVault,
  applyImagePullPolicy: applyImagePullPolicy,
  applySingleHostAffinity: applySingleHostAffinity,
  applyGoogleCloudServiceAccount: applyGoogleCloudServiceAccount,
  applyLivenessAndReadinessProbes: applyLivenessAndReadinessProbes,
  applyResourceConstraints: applyResourceConstraints,
}
