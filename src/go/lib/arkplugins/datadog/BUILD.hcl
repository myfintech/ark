package "ark_plugin_datadog" {
  description = <<-DESC
  This is a plugin that builds a Kuberentes manifest for deploying datadog
  DESC
}

target "build" "image" {
  repo = "gcr.io/managed-infrastructure/ark/plugins/datadog"

  dockerfile = templatefile("${package.path}/Dockerfile", {
    from = golang-1-16-alpine.build.image.url
    modules = root.build.go_modules.url
    package = package
  })

  source_files = [
    package.path,
    "${workspace.path}/src/go/lib",
    "${workspace.path}/src/jsonnet/lib",
  ]
}
