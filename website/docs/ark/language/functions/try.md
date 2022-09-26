---
layout: "functions"
page_title: "try - Functions - Configuration Language"
sidebar_current: "docs-funcs-conversion-try"
description: |-
  The try function tries to evaluate a sequence of expressions given as
  arguments and returns the result of the first one that does not produce
  any errors.
---

# `try` Function

`try` evaluates all of its argument expressions in turn and returns the result
of the first one that does not produce any errors.

This is a special function that is able to catch errors produced when evaluating
its arguments, which is particularly useful when working with complex data
structures whose shape is not well-known at implementation time.

For example, if some data is retrieved from an external system in JSON or YAML
format and then decoded, the result may have attributes that are not guaranteed
to be set. We can use `try` to produce a normalized data structure which has
a predictable type that can therefore be used more conveniently elsewhere in
the configuration:

```hcl
target "docker_exec" "test" {
  image = go.build.test_image.url
  working_directory = workspace.path
  command = [
    "go",
    "test",
    "-v",
    "./...",
  ]
  environment = {
    DOCKER_TLS_VERIFY = try(env.DOCKER_TLS_VERIFY, "")
    DOCKER_HOST = try(env.DOCKER_HOST, "")
    DOCKER_CERT_PATH = try(env.DOCKER_CERT_PATH, "")
    WATCHMAN_SOCK = try(env.WATCHMAN_SOCK, "")
  }
  volumes = [
    "${workspace.path}:${workspace.path}",
    "/docker/certs:/docker/certs",
  ]
  privileged = true
}
```

With the above environment map, a lookup of an environment variable that is not set
will not throw an error, instead setting the value in the map to the default
value, an empty string in this case.

We can also use `try` to deal with situations where a value might be provided
in two different forms, allowing us to normalize to the most general form:

We strongly suggest using `try` only in special arguments values whose expressions
perform normalization, so that the error handling is confined to a single
location in the target and the rest of the target can just use straightforward
references to the normalized structure and thus be more readable for future
maintainers.

The `try` function can only catch and handle _dynamic_ errors resulting from
access to data that isn't known until runtime. It will not catch errors
relating to expressions that can be proven to be invalid for any input, such
as a malformed target reference.

~> **Warning:** The `try` function is intended only for concise testing of the
presence of and types of object attributes. Although it can technically accept
any sort of expression, we recommend using it only with simple argument
references and type conversion functions.
Overuse of `try` to suppress errors will lead to a configuration that is hard
to understand and maintain.

## Examples

```
> local.foo
{
  "bar" = "baz"
}
> try(local.foo.bar, "fallback")
baz
> try(local.foo.boop, "fallback")
fallback
```

The `try` function will _not_ catch errors relating to constructs that are
provably invalid even before dynamic expression evaluation, such as a malformed
reference or a reference to a top-level object that has not been declared:

```
> try(local.nonexist, "fallback")

Error: Reference to undeclared local value

A local value with the name "nonexist" has not been declared.
```
