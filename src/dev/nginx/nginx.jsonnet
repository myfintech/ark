local ctx = std.native('ark_context')();
local microservice = import 'microservice/microservice.libsonnet';

local nginxMicroservice = microservice(
  name=ctx.name,
  image=ctx.image,
  command=['nginx', '-g', 'daemon off;'],
  replicas=1,
  env=[],
  port=10001,
  servicePort=10001,
);

[
  nginxMicroservice.deployment,
  nginxMicroservice.service,
]
