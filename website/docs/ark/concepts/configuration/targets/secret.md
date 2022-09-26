---
id: secret
title: secret
sidebar_label: secret
---

# `secret`

Creates a Kubernetes secret in a Kubernetes cluster based on current context. 


## Example Usage

### Secret from files
```hcl
target "secret" "sdm" {
  optional = true
  secret_name = "sdm-auth-files"
  namespace = "testing-1"
  files = ["${env.HOME}/.sdm"]
}
```

### Secret from environment variables
```hcl
target "secret" "sdm" {
  optional = true
  secret_name = "sdm-token"
  namespace = "testing-1"
  environment = ["SDM_SERVICE_TOKEN"]
}
```

## Attribute Reference

### Inputs

| Attribute | Required | Type | Explanation |
| --------- | :------: | ---- | ----------- |
| `optional` | :heavy_check_mark: | `boolean` | Whether or not the files or env vars are required. If they are not required and are not found, the creation of the resource for that file/var will be skipped. |
| `secret_name` | :heavy_check_mark: | `string` | The name of the resource. This is referenced in manifests and shown when running `kubectl get secrets`. |
| `namespace` |  | `string` | The namespace where the secret resource should be created. If this is not provided, the namespace configured in the loaded context will be used. |
| `files` |  | `array of strings` | One of `files` OR `environment` must be included in the target, but not both. A list of files or directories that should be read into the secret resource. |
| `environment` |  | `array of strings` | One of `files` OR `environment` must be included in the target, but not both. A list of environment variables that should have their values aggregated into the secret resource. |

### Outputs

This target has no outputs at this time.
