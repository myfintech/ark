---
id: docker_exec
title: Docker Exec
sidebar_label: docker_exec
---

# `docker_exec`

Executes a command within a docker container using the host system's vanilla Docker Engine (not Kubernetes or `docker-compose`).


## Example Usage

```hcl
target "docker_exec" "nginx-docker" {
  command = ["nginx", "-g", "daemon off;"]
  image = dev.build.nginx.url
  ports = ["10001:10001"]
  detach = true
  kill_timeout = "30s"
}
```

## Attribute Reference

### Inputs

| Attribute | Required | Type | Explanation |
| --------- | :------: | ---- | ----------- |
| `command` | :heavy_check_mark: | `array of strings` | A comma separated list of command arguments to run inside of the container. |
| `image` | :heavy_check_mark: | `string` | The source image in which the command will be executed. |
| `environment` |  | `map of strings` | A map of environment variables that should be available within the container. |
| `volumes` |  | `array of strings` | A list of volumes to mount from the host to the container. Each mount must be formatted as follows: `<host path>:<container path>`. |
| `working_directory` |  | `string` | The path in the container where the command execution should take place. |
| `ports` |  | `array of strings` | A list of ports that should be mapped from the container to the host. Each port mapping must be formatted as follows: `<host port>:<container port>`. |
| `detach` |  | `boolean` | If `detach` is set to `true`, `ark` will not follow the container logs, waiting for the container exit status. This is useful for long-running containers that don't need to be started and stopped often (NGINX proxying for StrongDM, for example). |
| `kill_timeout` |  | `string` | If a `docker_exec` run is cancelled (via a keyboard interrupt or another signal), `kill_timeout` specifies the duration of time the container has to gracefully exit before it's sent a SIGKILL. Valid time formats are specified as `<number><unit>`; for example, `10s` is ten seconds, `1h` is one hour. Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`. The default value if `kill_timeout` is not specified is `10s`. |
| `privileged` |  | `bool` | A boolean value to run the exec container in privileged mode. Be extremely careful using this and make sure your use case requires the container to run as privileged. Defaults to `false` |

### Outputs

There are presently no outputs available for the `docker_exec` target.
