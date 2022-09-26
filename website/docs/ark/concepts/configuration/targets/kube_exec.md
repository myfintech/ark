---
id: kube_exec
title: Kube Exec
sidebar_label: kube_exec
---

# `kube_exec`

Executes a command in a container in a Kubernetes pod.


## Example Usage

```hcl
target "kube_exec" "test" {
  resource_type = "ds"
  resource_name = "nginx-ingress"
  command = ["sh", "-c", "ls -lha"]
  get_pod_timeout = "10s"
}
```

## Attribute Reference

### Inputs

| Attribute | Required | Type | Explanation |
| --------- | :------: | ---- | ----------- |
| `resource_type` | :heavy_check_mark: | `string` | The type of Kubernetes resource where the pod can be found. Acceptable values include but are not limited to the following: `ds` for daemon sets, `deploy` for deployments, `pod` for pods, `statefulset` for stateful sets. |
| `resource_name` | :heavy_check_mark: | `string` | The name of the resource where the pod is located. |
| `command` | :heavy_check_mark: | `array of strings` | A comma separated list of command arguments to run inside of the container. |
| `container_name` |  | `string` | By default, if a container name is not specified, Kubernetes will attempt the command execution in the first container in the pod spec. If a specific container is needed for the `kube_exec`, consider explicitly setting `container_name` in the target definition. |
| `get_pod_timeout` |  | `string` | Limits the amount of time the Kubernetes API has to return a pod/container for connections. Valid time formats are specified as `<number><unit>`; for example, `10s` is ten seconds, `1h` is one hour. Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`. |

### Outputs

There are presently no outputs available for the `kube_exec` target. 
