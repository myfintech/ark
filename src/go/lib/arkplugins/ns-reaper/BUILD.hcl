package "ark-plugin-ns-reaper" {
  description = "Supplies a docker image that outputs a stateful set manifest for Postgres"
}

target "build" "image" {
  repo = "gcr.io/managed-infrastructure/ark/plugins/ns-reaper"

  dockerfile = templatefile("${package.path}/Dockerfile", {
    from = golang-1-16-alpine.build.image.url
    modules = root.build.go_modules.url
    package = package
  })

  source_files = [
    package.path,
    "${workspace.path}/src/go/lib/kube/microservice",
    "${workspace.path}/src/go/lib/kube/mutations",
    "${workspace.path}/src/go/lib/kube/objects",
    "${workspace.path}/src/go/lib/log",
  ]
}
