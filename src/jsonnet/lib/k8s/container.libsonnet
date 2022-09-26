function(
  name,
  image,
  command,
  env=[],
  args=[],
  ports=[],
  envFrom=[],
  volumeMounts=[],
  resources={},
  imagePullPolicy='IfNotPresent',
)
  {
    name: name,
    image: image,
    ports: ports,
    command: command,
    args: args,
    env: env,
    envFrom: envFrom,
    resources: resources,
    volumeMounts: volumeMounts,
    imagePullPolicy: imagePullPolicy,
  }
