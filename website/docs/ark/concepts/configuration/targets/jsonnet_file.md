---
id: jsonnet_file
title: Jsonnet File
sidebar_label: jsonnet_file
---

# `jsonnet_file`

Produces a jsonnet file. This target is particularly useful for producing YAML/JSON configuration manifests from a generic template (Kubernetes manifests, Buildkite pipelines, Docker Compose files, etc.)


## Example Usage

```hcl
target "jsonnet_file" "manifest" {
  format = "yaml"
  file = "./k8s/main.jsonnet"
  variables = jsonencode({
    image: buildkite.build.image.url
  })
  library_dir = [
    "${workspace.path}/src/jsonnet"
  ]
  source_files = [
    "./k8s"
  ]
}
```

## Attribute Reference

### Inputs

| Attribute | Required | Type | Explanation |
| --------- | :------: | ---- | ----------- |
| `file` | :heavy_check_mark: | `string` | The jsonnet template that should be used to render the jsonnet file |
| `variables` |  | `string` | Keys and value pairs represented as a single string. Consider using the `jsonencode()` function to create a map of variables; it makes life a lot easier. |
| `library_dir` |  | `array of strings` | A list of paths to (likely) generic jsonnet library files that can be referenced in package-specific jsonnet templates. |
| `format` |  | `string` | Determines the format that the rendered file will be. Accepts `json` or `yaml`. Default value is `json` if attribute is not explicity set.  |

### Outputs

| Attribute | Type | Explanation |
| --------- | ---- | ----------- |
| rendered_file | `string` | The filesystem path to the file that is rendered from the executed target. | 
