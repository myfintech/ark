---
id: exec
title: Local Exec
sidebar_label: exec
---

# `exec`

Runs a command on the host operating system.


## Example Usage

```hcl
target "exec" "install.tilt" {
  command = [
    "sh",
    "-c",
    "install -C ${dev.http_archive.tilt.contents_path}/tilt /usr/local/bin"
  ]
}
```

## Attribute Reference

### Inputs

| Attribute | Required | Type | Explanation |
| --------- | :------: | ---- | ----------- |
| `command` | :heavy_check_mark: | `array of strings` | A list of parameters that make up a command. |
| `environment` |  | `map of strings` | Key and value pairs of environment variables that should be set at the time of the command's execution. |

### Outputs

There are presently no outputs available for the `exec` target. 
