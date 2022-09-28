---
id: jsonnet
title: Jsonnet
sidebar_label: jsonnet
---

# `jsonnet`

Renders jsonnet into the specified format. This target is particularly useful for producing YAML/JSON configuration manifests from a generic template (Kubernetes manifests, Buildkite pipelines, Docker Compose files, etc.)


## Example Usage

```hcl
target "jsonnet" "manifest" {
  output_dir = artifact("output")
  library_dir = [
    "${workspace.path}/src/jsonnet"]
  yaml = true
  files = [
    "${package.path}/k8s/development.jsonnet",
    "${package.path}/k8s/integration.jsonnet",
    "${package.path}/k8s/uat.jsonnet",
    "${package.path}/k8s/production.jsonnet",
  ]

  source_files = [
    "${package.path}/k8s"
  ]

  variables = jsonencode({
    "name": package.name
  })
}
```

## Attribute Reference

### Inputs

| Attribute | Required | Type | Explanation |
| --------- | :------: | ---- | ----------- |
| `files` | :heavy_check_mark: | `array of strings` | The jsonnet template that should be used to render jsonnet files. |
| `output_dir` | :heavy_check_mark: | `string` | Is the path where the rendered jsonnet will be written. Typically you will use the artifact function to generate a target-specific artifact path.
| `variables` |  | `string` | Should be a json string value. These variables are accessible from jsonnet by using the function `std.native('ark_context')()`. PROTIP™: Consider using the `jsonencode()` function to render an HCL map as a json string. |
| `library_dir` |  | `array of strings` | A list of paths to jsonnet library files that can be referenced in package-specific jsonnet templates. |
| `yaml` |  | `bool` | A boolean value to set the output to yaml. Defaults to `false` which renders json. |

### Outputs

There are presently no outputs for the `jsonnet` target.
