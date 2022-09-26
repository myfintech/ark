package "ark-plugin-consul" {
  description = "Supplies a docker image that outputs a stateful set manifest for Consul"
}

target "build" "image" {
  repo = "gcr.io/managed-infrastructure/ark/plugins/consul"

  dockerfile = templatefile("${package.path}/Dockerfile", {
    from = golang-1-14-alpine.build.image.url
    modules = root.build.go_modules.url,
    package = package,
  })

  source_files = [
    package.path,
    "${workspace.path}/src/go/lib",
    "${workspace.path}/src/jsonnet/lib",
  ]
}
