---
id: build
title: Build
sidebar_label: build
---

# `build`

Executes a build in a sand-boxed container environment. Usage is meant for creating Docker images or compiling a binary.


## Example Usage

### Dockerfile as HEREDOC
```hcl
target "build" "nginx" {
  repo = "gcr.io/[insert-google-project]/mginx"
  dockerfile = <<-DOCKERFILE
  FROM nginx:alpine
  COPY . .
  RUN mv /src/dev/nginx/default.conf /etc/nginx/conf.d/default.conf
  RUN mv /src/dev/nginx/nginx.conf /etc/nginx/nginx.conf
  RUN --mount=type=secret,id=htpasswd cat /run/secrets/htpasswd | jq -rj '.password' >| /etc/nginx/.htpasswd
  DOCKERFILE

  secrets = {
    "htpasswd": "oao/development/bunchofnonsense/htpasswd"
  }

  source_files = [
    "${package.path}/nginx/default.conf",
    "${package.path}/nginx/nginx.conf"
  ]
}
```

### Dockerfile as template
```hcl
target "build" "nginx" {
  repo = "gcr.io/[insert-google-project]/nginx"
  dockerfile = templatefile("${package.path}/nginx/Dockerfile", {})
  source_files = [
      "${package.path}/nginx/default.conf",
      "${package.path}/nginx/nginx.conf"
    ]
}
```

## Attribute Reference

### Inputs

| Attribute | Required | Type | Explanation |
| --------- | :------: | ---- | ----------- |
| `repo` | :heavy_check_mark: | `string` | The name of the resulting Docker image |
| `dockerfile` | :heavy_check_mark: | `string` | The string representation of the contents of a Dockerfile. The recommended use of this attribute is either with HEREDOC syntax, creating a Dockerfile inline with the target, or using the HCL `templatefile()` function. Using `templatefile()` allows for a high degree of customization as Dockerfile contents can be templatized and provided at the time of the function call. |
| `build_args` |  | `map of strings` | A map of arguments and values required to successfully build the target image. |
| `target` |  | `string` | The Docker build phase to execute when named build phases are in place. |
| `tags` |  | `array of strings` | Any additional tags to create for a resulting Docker image. If an image is built, the default tags that are always created are `latest`, the deterministic build hash, and a short version of the build hash. |
| `output` |  | `string` | The path on the host filesystem where contents from within the Docker build are copied. This is extremely useful for obtaining something like a go binary. |
| `dump_context` |  | `boolean` | `dump_context` is a troubleshooting tool used for inspecting the contents of a build context as well as viewing the resulting Dockerfile from the `dockerfile` argument. Docker builds images using a g-zipped tarball of source files. `ark` requires the context to be explicitly defined, and it follows the `.gitignore`, which replicates the behavior of builds that take place within CI. This can result in files not making it to the context because they were left out of mistakenly assumed to be present when they're actually ignored. `dump_context` lets the user see *exactly* which files Docker is using for building. | 
| `cache_inline` |  | `boolean` | Enables the `buildkit` `BUILDKIT_INLINE_CACHE` build argument. Please reference [this][docker_caching] document for an explanation of how Docker uses this argument to make an image that can then be used as a cache source |
| `cache_from` |  | `array of strings` | Provides a list of images to pull and use as possible cache sources. **NOTE** Only images that contain layer metadata (please see the `cache_inline` argument for additional information on this) will be able to successfully be used as cache sources |
| `secrets` |  | `map of strings` | Keys and values must be formatted as follows: `<secret ID>:<vault path>`. The `secret ID` is how the secret is referenced in the `Dockerfile`, using the experimental `secret` [mount type][docker_secrets]. The Vault path value is the path in Vault to the secret. If you're using a secret mount, make sure your `Dockerfile` starts with this line: `# syntax = docker/dockerfile:1.0-experimental`. |
| `live_sync_enabled` |  | `boolean` | Enables real-time synchronization of file changes into Kubernetes containers. |
| `live_sync_force_rebuild_patterns` |  | `array of strings` | A list of file patterns that should trigger a complete rebuild of the image rather than synchronization. | 

### Outputs

| Attribute | Type | Explanation                                                                                                         |
| --------- | ---- |---------------------------------------------------------------------------------------------------------------------|
| `repo` | `string` | The base of an image URL. For example, `gcr.io/[insert-google-project]/[yourorg]/core-proxy`.                       |
| `url` | `string` | A complete image URL for a Docker image. The image URL is presented in the following format: `<repo>:<image hash>`. |
| `output` | `string` | A filesystem path to an output artifact when the `output` target attribute is provided.                             |

[docker_caching]: https://docs.docker.com/engine/reference/commandline/build/#specifying-external-cache-sources
[docker_secrets]: https://docs.docker.com/develop/develop-images/build_enhancements/#new-docker-build-secret-information
