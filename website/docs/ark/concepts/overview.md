---
id: overview
title: Overview
sidebar_label: Overview
---

## Ark HCL Files

HCL (HashiCorp Configuration Language) is a configuration language built by [HashiCorp](https://www.hashicorp.com/). The goal of HCL is to build a structured configuration language that is both human and machine-friendly for use with command-line tools but specifically targeted towards DevOps tools, servers, etc. This is the language of choice for tools like [Terraform](https://www.terraform.io/), [Vault](https://www.vaultproject.io/), and [Docker BuildX](https://github.com/docker/buildx/pull/192).

***Ark uses the [go HCL AST](https://pkg.go.dev/github.com/hashicorp/hcl/v2@v2.4.0/gohcl) (abstract syntax tree) parser to provide a DSL (domain-specific language) for building docker images and defining developer-centric workflows against kubernetes.***

If you want to read more about HCL you can checkout [their github page](https://github.com/hashicorp/hcl).

We will provide more examples of the syntax in a later version of this guide.

- Example

## Workspaces

A `workspace` represents a position is your file system hierarchy which is the root of your project. Workspaces can be nested to create boundaries between projects. Ark traverses your filesystem from the current working directory up to locate your `WORKSPACE.hcl`. The directory where this file lives is your `workspace root`. All `packages` are loaded relative to the `workspace root`.

## Packages

A `package` is a container for build, task, and workflow targets. All `*.ark.hcl` files must declare the package they belong to.

```hcl
package "core-proxy" {
	description = ""
}
```

## Targets

A `target` can be thought of as an achievable state or goal; examples of targets might be a workflow or build to execute. Targets are referenced by their `address`; examples of target addresses are `core-proxy.build.server` (which builds the server image) and `core-proxy.workflow.migrate` (which executes the migration workflow).

### build

A `build` target is used to produce a docker image. An image can be used to deploy an application or to cache some kind of artifact for another build step. Linking multiple build targets together can allow you to create more incremental builds and only rebuild what has changed.

There are several benefits to executing your builds in docker

- **Hermeticity**

    The most important benefit you gain from using the `build` target is hermeticity. Your build steps are executed from the context of an isolated docker build stage. This almost eliminates the highest source of entropy to your build process (your host file system). We achieve this through encapsulating and pinning build and runtime dependencies in docker layers, instead of relying on host OS-specific versions and changing installation processes. This creates parody between local, remote, and CI environments, and helps us reduce the frequency of the phrase `it worked on my machine`.

    The only exception to this is the docker build context. This `context` is a snapshot of part of your local file system that is submitted to the Docker daemon when executing a build. There are known limitations of `.dockerignore` that can still produce inconsistent behavior between local and CI-based (fresh clone) builds.

    [We are actively working on a holistic approach to this problem](https://www.notion.so/mantl/Custom-Docker-Frontend-Buildkit-Integration-0fac5920ffc04806a82dceacda2d0955), but to provide incremental value we have opted to deliver the first iteration of this tool without a custom docker front end that would allow us to more heavily control the file system synchronization between local and remote builds.

- Local and remote artifact caching

    The docker image registry has evolved into an new standard called the OCI Artifact Registry. If it can be stored in an image, it can be cached locally and remotely in an artifact registry. Assuming a CI system has built and remotely cached any branch, you can pull a copy of all matching artifacts so you can avoid building from scratch (saving massive amounts of time and CPU cycles).

#### `BUILD.hcl`

```hcl
package "root" {
  description = "go modules or node_modules go here"
}

target "build" "node_modules" {
    context = workspace.path
    dockerfile = file("${workspace.path}/src/docker/node_modules.docker")
    source_files = ["./package.json", "./package-lock.json"]
}
```

- `context` is the same as a [docker context][docker_context] but ark has special behavior that enhances how contexts behave
- `dockerfile` is a string representation of a `Dockerfile`. Here we are using the ark std library `file` function to load the contents of a `Dockerfile` in another directory.
- `source_files` are one of the most important concepts in ark. This defines ALL the files & directories that your target relies on. The more explicit you are the better ark will be able to cache the artifacts of your builds. Ark uses the `source_files` as input to generate a deterministic hash of each target. If nothing has changed in a source, the hash will remain the same.  


#### `src/docker/node_modules.docker`

```docker
FROM node:10-alpine
WORKDIR /opt/deps
COPY ./package.json ./package.json
COPY ./yarn.lock ./yarn.lock
```

- The contents of this file is standard docker syntax.
- This is how ark differentiates itself from other tools like [Uber Makisu][makisu] or [Google Kaniko][kaniko].
- We retain full compatibility with Dockerfile syntax but extend it to support a templating language making them more flexible and modular.

#### `src/ts/services/client-service/BUILD.hcl`

```hcl
package "client-microservice" {
  description = "The client micro service package"
}

target "build" "image" {
    context = workspace.path
    dockerfile = templatefile("${package.path}/Dockerfile", {
      node_modules = root.build.node_modules.url
    })
}
```

- You'll notice this time we used the `temnplatefile` function. This allows us to create `Dockerfiles` with an extended syntax.
- Docker already supports `BUILD_ARGS` but there are common pitfalls that make them a bad option. (TODO: add build arg pitfalls) 

#### `src/ts/services/client-service/Dockerfile`

```docker
ARG node_modules
FROM ${node_modules} as deps

FROM node:10-alpine
WORKDIR /opt/app
COPY --from=deps /opt/deps /opt/app
```



[makisu]: https://eng.uber.com/makisu/
[kaniko]: https://github.com/GoogleContainerTools/kaniko
[docker_context]: https://docs.docker.com/engine/reference/commandline/build/#tarball-contexts