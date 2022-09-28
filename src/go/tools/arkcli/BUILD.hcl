package "arkcli" {
  description = <<-EOT
  The arkcli is the user interface that's backed by the arksdk.
  It provides various commands to its users to easily and
  intuitively build their tools and services.
  EOT
  version = "0.1.1"
}

locals {
  environment = "development"
  upload_url  = "gs://[cdn-domain]/assets/arkcli"
  http_url    = "https://[cdn-domain]/assets/arkcli"
}

target "build" "bin" {
  repo = "gcr.io/[insert-google-project]/domain/${package.name}"
  dockerfile = templatefile("${workspace.path}/src/docker/go_tools.docker", {
    from                     = golang-1-16-alpine.build.image.url
    package                  = package
    latest_version_base_url  = locals.http_url
    latest_download_base_url = locals.http_url
    modules                  = root.build.go_modules.url
    environment              = try(env.APP_ENV, locals.environment)
  })

  source_files = [
    "${package.path}/cmd",
    "${package.path}/main.go",
    "${package.path}/version.go",
    "${workspace.path}/src/go/lib/log",
    "${workspace.path}/src/go/lib/pkg",
    "${workspace.path}/src/go/lib/arksdk",
    "${workspace.path}/src/go/lib/jsonnetutils",
    "${workspace.path}/src/go/lib/exec",
    "${workspace.path}/src/go/lib/ark",
    "${workspace.path}/src/go/lib/iomux",
    "${workspace.path}/src/go/lib/fs",
    "${workspace.path}/src/go/lib/hclutils",
    "${workspace.path}/src/go/lib/git/gitignore",
    "${workspace.path}/src/go/lib/container",
    "${workspace.path}/src/jsonnet/lib",
    "${workspace.path}/src/go/lib/utils",
    "${workspace.path}/src/go/lib/pattern",
    "${workspace.path}/src/go/lib/kube",
    "${workspace.path}/src/go/lib/watchman",
    "${workspace.path}/src/go/lib/net",
    "${workspace.path}/src/go/lib/dag",
    "${workspace.path}/src/go/lib/state_store",
    "${workspace.path}/src/go/lib/grpc",
    "${workspace.path}/src/go/lib/vault_tools",
    "${workspace.path}/src/go/lib/internal_net",
    "${workspace.path}/src/go/lib/xdgbase",
    "${workspace.path}/src/go/lib/gorm/json_datatypes",
    "${workspace.path}/src/go/lib/embedded_scripting/typescript",
  ]

  output = artifact("bin")
}

target "local_file" "install_script" {
  filename = "install.sh"
  content = templatefile("${package.path}/install", {
    download_url = "${locals.http_url}/versions/${arkcli.build.bin.hash}/bin"
  })
  source_files = [
  "${package.path}/install"]
}

target "docker_exec" "upload" {
  image             = "google/cloud-sdk:297.0.1-alpine"
  working_directory = "/opt/ark"
  volumes = [
    "${env.HOME}/.config/gcloud:/root/.config/gcloud",
    "${arkcli.local_file.install_script.artifacts.path}:/opt/ark/install",
    "${arkcli.build.bin.artifacts.path}/bin:/opt/ark/bin"
  ]

  command = [
    "sh",
    "-c",
    <<-INSTALL
    echo ${locals.upload_url}/versions/${arkcli.build.bin.hash}/ && \
    gsutil -m -h "Cache-Control:no-cache,max-age=0" \
      cp -r ./ ${locals.upload_url}/versions/${arkcli.build.bin.hash}/
    INSTALL
  ]
}

target "docker_exec" "publish" {
  image             = "google/cloud-sdk:297.0.1-alpine"
  working_directory = "/opt/ark"
  volumes = [
    "${env.HOME}/.config/gcloud:/root/.config/gcloud",
    "${arkcli.local_file.install_script.artifacts.path}:/opt/ark/install",
    "${arkcli.build.bin.artifacts.path}/bin:/opt/ark/bin"
  ]

  command = [
    "sh",
    "-c",
    <<-INSTALL
    echo ${locals.upload_url}/versions/${arkcli.build.bin.hash}/ && \
    gsutil -h "Cache-Control:no-cache,max-age=0" \
      cp -r ./install/install.sh ${locals.upload_url}/install.sh && \

    gsutil -m -h "Cache-Control:no-cache,max-age=0" \
      cp /opt/ark/bin/arkcli-* ${locals.upload_url}/
    INSTALL
  ]

  depends_on = ["arkcli.docker_exec.upload"]
}
