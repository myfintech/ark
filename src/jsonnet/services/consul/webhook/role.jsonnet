local roleBinding = import 'lib/k8s/role-binding.libsonnet';
local role = import 'lib/k8s/role.libsonnet';

function() {
  webhookRole: role(
    name='consul-connect-injector-webhook',
    labels={ app: 'consul' },
    kind='ClusterRole',
    rules=[
      {
        apiGroups: ['admissionregistration.k8s.io'],
        resources: ['mutatingwebhookconfigurations'],
        verbs: [
          'get',
          'list',
          'watch',
          'patch',
        ],
      },
    ],
  ),

  webhookRoleBinding: roleBinding(
    name='consul',
    labels={ app: 'consul' },
    kind='ClusterRoleBinding',
    subjects=[{
      kind: 'ServiceAccount',
      name: 'consul-connect-injector-webhook-svc-account',
      apiGroup: '',
      namespace: 'consul',
    }],
    roleRef={
      kind: 'ClusterRole',
      name: 'consul-connect-injector-webhook',
      apiGroup: 'rbac.authorization.k8s.io',
    }
  ),
}
