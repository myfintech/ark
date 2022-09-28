package "ark-entrypoint" {
  description = <<-EOT
  ark-entrypoint ...
EOT
}

target "build" "bin" {
  repo = "gcr.io/[insert-google-project]/domain/${package.name}"
  dockerfile = templatefile("${workspace.path}/src/docker/go_tools.docker", {
    from                     = golang-1-14-alpine.build.image.url
    package                  = package
    latest_version_base_url  = ""
    latest_download_base_url = ""
    modules                  = root.build.go_modules.url
    environment              = try(env.APP_ENV, "dev")
  })

  source_files = [
    "${package.path}/main.go",
    "${workspace.path}/src/go/lib/fs",
    "${workspace.path}/src/go/lib/log",
    "${workspace.path}/src/go/lib/exec",
    "${workspace.path}/src/go/lib/kube",
    "${workspace.path}/src/go/lib/utils",
    "${workspace.path}/src/go/lib/pattern",
    "${workspace.path}/src/go/lib/grpc",
    "${workspace.path}/src/go/lib/iomux",
    "${workspace.path}/src/go/lib/watchman",
    "${workspace.path}/src/go/lib/container",
    "${workspace.path}/src/go/lib/ark/entrypoint",
    "${workspace.path}/src/go/lib/ark/log_sink_server",
  ]

  depends_on = ["ark-entrypoint-lib.exec.protoc"]
}
