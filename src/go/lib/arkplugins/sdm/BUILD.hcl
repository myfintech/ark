package "sdm_plugin" {
  description = "Supplies a docker image that outputs a deployment manifest for SDM"
}

target "build" "sdm" {
  repo = "gcr.io/managed-infrastructure/ark/sdm"
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
