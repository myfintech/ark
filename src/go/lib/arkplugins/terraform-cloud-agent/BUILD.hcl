package "ark_plugin_terraform_cloud_agent" {
  description = <<-DESC
  This is a plugin that builds a Kubernetes manifest for deploying a terraform cloud agent
  DESC
}

target "build" "image" {
  repo = "gcr.io/managed-infrastructure/ark/plugins/terraform-cloud-agent"

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
