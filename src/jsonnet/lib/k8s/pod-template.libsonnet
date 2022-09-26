function(
  containers,
  nodeSelector={},
  restartPolicy='Always',
  serviceAccountName='',
  initContainers=[],
  hostNetwork=false,
  volumes=[],
  annotations={}
) {
  metadata: {
    annotations: annotations,
  },
  spec: {
    containers: containers,
    serviceAccountName: serviceAccountName,
    restartPolicy: restartPolicy,
    terminationGracePeriodSeconds: 30,
    hostNetwork: hostNetwork,
    initContainers: initContainers,
    volumes: volumes,
    nodeSelector: nodeSelector,
    imagePullSecrets: [],
    affinity: {},
  },
}
