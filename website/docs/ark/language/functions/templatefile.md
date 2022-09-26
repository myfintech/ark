---
layout: "functions"
page_title: "templatefile - Functions - Configuration Language"
sidebar_current: "docs-funcs-file-templatefile"
description: |-
  The templatefile function reads the file at the given path and renders its
  content as a template.
---

# `templatefile` Function

`templatefile` reads the file at the given path and renders its content
as a template using a supplied set of template variables.

```hcl
templatefile(path, vars)
```

The template syntax is the same as for
[string templates](../expressions.md#string-templates) in HCL, 
including interpolation sequences delimited with `${` ... `}`.
This function just allows longer template sequences to be factored out
into a separate file for readability.

The "vars" argument must be a map. Within the template file, each of the keys
in the map is available as a variable for interpolation. The template may
also use any other function available in Ark, except that
recursive calls to `templatefile` are not permitted. Variable names must
each start with a letter, followed by zero or more letters, digits, or
underscores.

Strings in HCL are sequences of Unicode characters, so
this function will interpret the file contents as UTF-8 encoded text and
return the resulting Unicode characters. If the file contains invalid UTF-8
sequences then this function will produce an error.

This function can be used only with files that already exist on disk at the
beginning of an Ark run. Functions do not participate in the dependency
graph, so this function cannot be used with files that are generated
dynamically during an Ark operation. We do not recommend using dynamic
templates in Ark configurations, but in rare situations where this is
necessary you can use
[the `local_file` target](../../concepts/configuration/targets/local_file.md)
to render templates while respecting target dependencies.

## Examples

Given a template file `backends.tmpl` with the following content:

```
%{ for addr in ip_addrs ~}
backend ${addr}:${port}
%{ endfor ~}
```

The `templatefile` function renders the template:

```
> templatefile("${path.module}/backends.tmpl", { port = 8080, ip_addrs = ["10.0.0.1", "10.0.0.2"] })
backend 10.0.0.1:8080
backend 10.0.0.2:8080

```

### Generating JSON or YAML from a template

If the string you want to generate will be in JSON or YAML syntax, it's
often tricky and tedious to write a template that will generate valid JSON or
YAML that will be interpreted correctly when using lots of individual
interpolation sequences and directives.

Instead, you can write a template that consists only of a single interpolated
call to either [`jsonencode`](./jsonencode.md) or
[`yamlencode`](./yamlencode.md), specifying the value to encode using
[normal HCL expression syntax](../expressions.md)
as in the following examples:

```
${jsonencode({
  "backends": [for addr in ip_addrs : "${addr}:${port}"],
})}
```

```
${yamlencode({
  "backends": [for addr in ip_addrs : "${addr}:${port}"],
})}
```

Given the same input as the `backends.tmpl` example in the previous section,
this will produce a valid JSON or YAML representation of the given data
structure, without the need to manually handle escaping or delimiters.
In the latest examples above, the repetition based on elements of `ip_addrs` is
achieved by using a
[`for` expression](../expressions.md#for-expressions)
rather than by using
[template directives](../expressions.md#directives).

```json
{"backends":["10.0.0.1:8080","10.0.0.2:8080"]}
```

If the resulting template is small, you can choose instead to write
`jsonencode` or `yamlencode` calls inline in your main configuration files, and
avoid creating separate template files at all.

For more information, see the main documentation for
[`jsonencode`](./jsonencode.md) and [`yamlencode`](./yamlencode.md).

## Related Functions

* [`file`](./file.md) reads a file from disk and returns its literal contents
  without any template interpretation.
