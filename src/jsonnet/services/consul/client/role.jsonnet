local roleBinding = import 'lib/k8s/role-binding.libsonnet';
local role = import 'lib/k8s/role.libsonnet';

function() {
  clientRole: role(
    name='consul-client',
    labels={ app: 'consul' },
    kind='ClusterRole',
    rules=[],
  ),

  clientRoleBinding: roleBinding(
    name='consul-client',
    labels={ app: 'consul' },
    kind='ClusterRoleBinding',
    subjects=[{
      kind: 'ServiceAccount',
      name: 'consul-client-svc-account',
      apiGroup: '',
      namespace: 'consul',
    }],
    roleRef={
      kind: 'ClusterRole',
      name: 'consul-client',
      apiGroup: 'rbac.authorization.k8s.io',
    }
  ),
}
