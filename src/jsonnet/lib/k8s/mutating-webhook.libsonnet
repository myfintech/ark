function(
  name,
  labels={},
  clientConfig={},
) {
  apiVersion: 'admissionregistration.k8s.io/v1beta1',
  kind: 'MutatingWebhookConfiguration',
  metadata: {
    name: name,
    labels: labels,
  },
  webhooks: [
    {
      clientConfig: clientConfig,
    },
  ],
}
