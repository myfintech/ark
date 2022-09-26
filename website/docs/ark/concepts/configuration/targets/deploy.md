---
id: deploy
title: deploy
sidebar_label: deploy
---

# `deploy`

Executes the equivalent of a `kubectl apply -f` followed by a `kubectl rollout status` to track the rollout of any observable resources (Deployments, DaemonSets, and StatefulSets).

## Example Usage

```hcl
target "deploy" "nginx" {
  manifest = manifest_plugin({name = "nginx"})
  env = [{
    name = "EXAMPLE_VAR"
    value = "example_value"
  }]

  live_sync_enabled = true
  live_sync_restart_mode = "delegated"
  live_sync_on_actions = [{
    command = ["bash", "-c", "nginx -t && nginx -s reload"]
    work_dir = "./"
    patterns = ["**/*.conf"]
  }]

  jsonnet_mutation_hook = "${package.path}/hook.jsonnet"
}
```

## Attribute Reference

### Inputs

| Attribute | Required | Type | Explanation |
| --------- | :------: | ---- | ----------- |
| `manifest` | :heavy_check_mark: | `string` | The string representation of the contents of a Kubernetes manifest. This particular field works best with either templates or via use of ark's plugin system. |
| `port_forward` |  | `array of strings` | Formatted as `<host port>:<container port>`, the `port_forward` attribute informs the `ark run` command with the `--watch` flag how service ports should be forwarded from Kubernetes to the host. |
| `live_sync_enabled` |  | `boolean` | Enables real-time synchronization of file changes into Kubernetes containers. |
| `live_sync_restart_mode` |  | Defaults to `auto`. Accepts `auto` or `delegated`. If live synchronization is enabled, `live_sync_restart_mode` informs the container synchronization server how process restarts should be handled. If `delegated` is chosen, the container's application process is responsible for restarting itself on file changes (think `nodemon`). If `auto` is chosen, the synchronization server will restart the process on file changes. |
| `live_sync_on_actions` |  | array of objects | See below for the `live_sync_on_actions` object definition. Informs the syncronization server if a command should be executed before the process restart takes place (installing node modules from an updated `package.json` for example). |
| `env` |  | `map of strings` | A map of environment variables that should be available to the container running the provided image. |

#### `live_sync_on_actions` Object Reference
**Note**: `live_sync_on_actions` is an array of these objects.

| Field | Type | Explanation |
| ----- | ---- | ----------- |
| `command` | `array of strings` | A comma-separated list of one or more commands to be executed. (example: `["yarn", "start"]` or <code>["bash", "-c", "ls -lah &#124; grep foo"]</code>) |
| `work_dir` | `string` | The directory from which the command(s) should be executed. |
| `patterns` | `array of strings` | File name patterns for which this action should be applied. |

### Outputs

| Attribute | Type | Explanation |
| --------- | ---- | ----------- |
| `rendered_file` | `string` | The location on disk where the rendered Kubernetes manifest is stored. |
