local roleBinding = import 'lib/k8s/role-binding.libsonnet';
local role = import 'lib/k8s/role.libsonnet';

function() {
  serverRole: role(
    name='consul-server',
    labels={ app: 'consul' },
    kind='ClusterRole',
    rules=[],
  ),

  serverRoleBinding: roleBinding(
    name='consul-server',
    labels={ app: 'consul' },
    kind='ClusterRoleBinding',
    subjects=[{
      kind: 'ServiceAccount',
      name: 'consul-server',
      apiGroup: '',
      namespace: 'consul',
    }],
    roleRef={
      kind: 'ClusterRole',
      name: 'consul-server-svc-account',
      apiGroup: 'rbac.authorization.k8s.io',
    }
  ),
}
