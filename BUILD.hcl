package "root" {
  description = ""
}

target "build" "go_modules" {
  repo = "gcr.io/[insert-google-project]/go-modules"

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
