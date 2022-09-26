local ms = import 'mantl/sre/microservice/microservice.libsonnet';

[ms(name='example', image='test', command=['sh'])]
