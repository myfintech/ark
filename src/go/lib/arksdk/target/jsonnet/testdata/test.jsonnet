local test = import 'test.libsonnet';
local ctx = std.native('ark_context')();
[ctx, test, {
  containers: [4, 5],
  env: {
    something: 'arbitrary',
  },
} {
  containers+: [1, 2, 3],
  env+: {
    anything: 'else',
  },
}]
