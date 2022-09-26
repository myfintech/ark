package "ark-plugin-vault-k8s-sa" {
  description = "A plugin that outputs a manifest for a service account and a cluster role binding to be used with Hashicorp Vault"
}

target "build" "image" {
  repo = "gcr.io/managed-infrastructure/ark/plugins/vault-k8s-sa"

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
