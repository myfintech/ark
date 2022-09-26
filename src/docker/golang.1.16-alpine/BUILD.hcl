package "golang-1-16-alpine" {
  description = "A CLI tool to generate a sha256 hash for an input stdin, file, or dir"
}

target "build" "image" {
  repo = "gcr.io/managed-infrastructure/mantl/${package.name}"

  dockerfile = <<-DOCKERFILE
  FROM ${checksum.build.bin.url} as checksum
  FROM ${pkginfo.build.bin.url} as pkginfo
  FROM golang:1.16-alpine as build
  COPY --from=checksum /checksum-linux /bin/checksum
  COPY --from=pkginfo /pkginfo-linux /bin/pkginfo
  DOCKERFILE
}
