package "ark_plugin_kube_state" {
  description = <<-DESC
  This is a plugin that builds a Kuberentes manifest for deploying kube-state-metrics
  DESC
}

target "build" "image" {
  repo = "gcr.io/managed-infrastructure/ark/plugins/kube-state"

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
