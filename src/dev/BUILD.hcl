package "dev" {
  description = ""
}

locals {
  operator_name = "confluent-operator-1.0"
}

target "build" "kube_deps" {
  repo = "gcr.io/[insert-google-project]/${package.name}"

  dockerfile = templatefile("${workspace.path}/src/docker/dev_base.docker", {
    from = "alpine"
    run_commands = [
      "apk add --no-cache ca-certificates curl openssl ncurses bash",
      "curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl",
      "chmod +x ./kubectl",
      "mv ./kubectl /usr/local/bin/kubectl",
      "curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | sh"
    ]
  })
}

// Confluent things
target "http_archive" "confluent_operator" {
  url    = "https://[cdn-domain]/ark/external/confluent-operator/${locals.operator_name}.tar.gz"
  sha256 = "8031142b440bd41dd480f585954aeee90599f759ecf2d42acc2d69a1bfdf7d35"
}

target "docker_exec" "install_confluent_operator" {
  working_directory = "/opt/install"
  volumes = [
    "${dev.http_archive.confluent_operator.contents_path}/${locals.operator_name}.tar.gz:/opt/install/${locals.operator_name}.tar.gz",
    "${env.HOME}/.kube/config:/root/.kube/config"
  ]
  image = dev.build.kube_deps.url
  command = [
    "sh",
    "-c",
    <<-ICO
    tar -xf /opt/install/${locals.operator_name}.tar.gz
    cd ${locals.operator_name}
    bash ./scripts/operator-util.sh \
    -n confluent \
    -r ${locals.operator_name} \
    -f helm/providers/private.yaml
    ICO
  ]
}

// Tilt things
target "http_archive" "tilt" {
  url        = "https://github.com/tilt-dev/tilt/releases/download/v0.14.2/tilt.0.14.2.mac.x86_64.tar.gz"
  sha256     = "72608731c4599f773d8623740dae526a5209fd853d2223103034f3065bc70966"
  decompress = true
}

target "exec" "install.tilt" {
  command = [
    "sh",
    "-c",
    "install -C ${dev.http_archive.tilt.contents_path}/tilt /usr/local/bin"
  ]
}

// Postgres things
target "docker_exec" "install_postgres" {
  working_directory = "/opt/install"
  volumes = [
    "${env.HOME}/.kube/config:/root/.kube/config"
  ]
  image = dev.build.kube_deps.url
  command = [
    "sh",
    "-c",
    <<-ICO
    helm repo add bitnami https://charts.bitnami.com/bitnami
    helm install core-proxy \
    --set global.postgresql.postgresqlPassword=user \
    --set postgresqlPassword=password \
    --set persistence.mountPath=/ark \
    bitnami/postgresql
    ICO
  ]
}

// NGINX things
target "build" "nginx" {
  repo         = "gcr.io/[insert-google-project]/nginx"
  dockerfile   = templatefile("${package.path}/nginx/Dockerfile", {})
  dump_context = true
  source_files = [
    "${package.path}/nginx/default.conf",
    "${package.path}/nginx/nginx.conf"
  ]
}

target "jsonnet_file" "nginx" {
  format      = "yaml"
  library_dir = ["${workspace.path}/src/jsonnet"]
  file        = "${package.path}/nginx/nginx.jsonnet"
  variables = jsonencode({
    "name" : "dev-nginx",
    "image" : dev.build.nginx.url
  })
}

target "exec" "nginx_kubectl" {
  command = ["kubectl", "apply", "-f", dev.jsonnet_file.nginx.rendered_file]
}

target "docker_exec" "nginx_docker" {
  command = ["nginx", "-g", "daemon off;"]
  image   = dev.build.nginx.url
  ports   = ["10001:10001"]
  detach  = true
}

target "deploy" "gcloud_emulator" {
  manifest = gcloud-emulator({
    name = "pubsub-emulator",
    emulator = "pubsub",
    project = "test"
  })
}

target "deploy" "postgres" {
  manifest = postgres({
    name = "postgres"
  })
}

target "deploy" "vault" {
  manifest = vault({
    name = "vault"
  })
  port_forward = [
    "8200:8200"
  ]
}

//target "kv_sync" "seed_vault" {
//  engine = "vault"
//  engine_url = "http://localhost:8200"
//  timeout = "30s"
//  token = "root"
//
//  source_files = [
//    "${workspace.path}/.ark/kv/foo",
//    "${workspace.path}/.ark/kv/bar",
//    "${workspace.path}/.ark/kv/fruit"
//  ]
//
//  depends_on = [
//    "dev.deploy.vault",
//  ]
//}
