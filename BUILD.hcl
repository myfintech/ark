package "root" {
  description = <<-EOT
  The root of the MANTL repo workspace. From this package,
  golang and node modules can be downloaded for use in the
  various services, libraries, and tools in this glorious repo.
  EOT
}

target "build" "go_modules" {
  repo = "gcr.io/managed-infrastructure/mantl/go-modules"

  dockerfile = templatefile("${package.path}/src/docker/modules.docker", {
    from = "golang:1.16-alpine"
    commands = [
      "apk add --no-cache git",
      "go mod download",
    ]
  })

  source_files = [
    "${package.path}/go.mod",
    "${package.path}/go.sum",
  ]
}
