package "ark_plugin_core_proxy" {
  description = <<-DESC
  This is a plugin that builds a Kuberentes manifest for deploying core-proxy
  DESC
}

target "build" "image" {
  repo = "gcr.io/managed-infrastructure/ark/plugins/core-proxy"

  dockerfile = templatefile("${package.path}/Dockerfile", {
    from = golang-1-14-alpine.build.image.url
    modules = root.build.go_modules.url
    package = package
  })

  source_files = [
    package.path,
    "${workspace.path}/src/go/lib",
    "${workspace.path}/src/jsonnet/lib",
  ]
}