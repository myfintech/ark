local podDisruptionBudget = import 'lib/k8s/pod-disruption-budget.libsonnet';

function() {
  disruptionBudget: podDisruptionBudget(
    name='consul-server',
    labels={ app: 'consul' },
    matchLabels={
      app: 'consul',
      component: 'server',
    },
  ),
}
