---
id: local_file
title: local_file
sidebar_label: local_file
---

# `local_file`

Produces a file on a local filesystem. 


## Example Usage

```hcl
target "local_file" "tilt" {
  filename = "${package.path}/tilt.json"
  content = jsonencode({
    service_files = [
      core-proxy.local_file.tilt.filename,
    ]
    workspace = workspace
  })
}
```

## Attribute Reference

### Inputs

| Attribute | Required | Type | Explanation |
| --------- | :------: | ---- | ----------- |
| `filename` | :heavy_check_mark: | `string` | The filesystem path where the created file will be placed. |
| `content` | :heavy_check_mark: | `string` | The file's content. In the target example, a JSON-encoded string is produced that consists of keys/values from other available inputs. |
| `file_permissions` |  | `octal mode value` | Defaults to `0644`, which equates to the symbolic reference `-rw-r--r--`, allowing the user read and write access, the group read access, and everyone read access. |
| `directory_permissions` |  | `octal mode value` | Defaults to `0755`, which equates to the symbolic reference `-rwxr-xr-x`, allowing the user read, write, and execute access; the group read and execute access; and everyone read and execute access. |

### Outputs

| Attribute | Type | Explanation |
| --------- | ---- | ----------- |
| `filename` | `string` | A direct output of the `filename` input attribute if the path needs to be referenced in another target. | 
