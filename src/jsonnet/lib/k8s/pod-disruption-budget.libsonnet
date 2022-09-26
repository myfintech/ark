function(
  name,
  maxUnavailable=0,
  labels={},
  selector={},
  matchLabels={},
) {
  apiVersion: 'policy/v1beta1',
  kind: 'PodDisruptionBudget',
  metadata: {
    name: name,
    labels: labels,
  },
  spec: {
    maxUnavailable: maxUnavailable,
    selector: selector {
      matchLabels: matchLabels,
    },
  },
}
