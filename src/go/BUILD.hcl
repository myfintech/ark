package "go" {
  description = "The home for all MANTL golang source code."
}

target "build" "test_image" {
  repo = "gcr.io/managed-infrastructure/mantl/go/test"
  // language=docker
  dockerfile = <<-DOC
  FROM docker:latest as docker
  FROM gcr.io/managed-infrastructure/mantl/gcloud:latest as google-cloud-sdk
  FROM ${root.build.go_modules.url} as test
  RUN apk update && apk upgrade
  RUN apk add --no-cache --update ca-certificates openssl openssl openssl-dbg openssl-dev
  RUN apk add --no-cache \
  autoconf \
  automake \
  bash \
  build-base \
  curl \
  docker \
  git \
  gnupg \
  jq \
  libc6-compat \
  libcrypto1.1 \
  libgcc \
  libstdc++ \
  libtool \
  linux-headers \
  musl-dev \
  openssh-client \
  openssl-dev \
  perl \
  py-pip \
  python3-dev \
  rsync \
  run-parts \
  su-exec \
  tini \
  tzdata \
  unzip \
  wget
  COPY --from=docker /usr/local/bin/docker /usr/local/bin/docker
  COPY --from=google-cloud-sdk /root/google-cloud-sdk /opt/google-cloud-sdk
  RUN /opt/google-cloud-sdk/bin/gcloud config set --installation component_manager/disable_update_check true
  ENV PATH /opt/google-cloud-sdk/bin:$PATH
  RUN gcloud auth configure-docker
  DOC
}

target "docker_exec" "test" {
  image             = go.build.test_image.url
  working_directory = workspace.path
  command = [
    "bash",
    "-c",
    <<-script
    set -euo pipefail
    gcloud auth activate-service-account --key-file=$${GOOGLE_APPLICATION_CREDENTIALS}
    gcloud auth configure-docker --quiet
    gcloud config set project managed-infrastructure
    kubectl create namespace $${ARK_K8S_NAMESPACE}
    set +e
    unset CGO_ENABLED
    go get -tags musl github.com/confluentinc/confluent-kafka-go/kafka
    go mod vendor
    go test -tags musl -v ./...
    exitCode="$${?}"
    set -e
    kubectl delete namespace $${ARK_K8S_NAMESPACE}
    exit "$${exitCode}"
    script
  ]
  environment = {
    DOCKER_TLS_VERIFY              = try(env.DOCKER_TLS_VERIFY, "")
    DOCKER_HOST                    = try(env.DOCKER_HOST, "")
    DOCKER_CERT_PATH               = try(env.DOCKER_CERT_PATH, "")
    WATCHMAN_SOCK                  = try(env.WATCHMAN_SOCK, "")
    VAULT_ADDR                     = try(env.VAULT_ADDR, "https://vault.mantl.team")
    ARK_K8S_SAFE_CONTEXTS          = try(env.ARK_K8S_SAFE_CONTEXTS, "development_admin,development_sre,gke_managed-infrastructure_us-central1_development-us-central1")
    ARK_K8S_NAMESPACE              = "buildkite-${try(env.BUILDKITE_BUILD_ID, "")}"
    KUBECONFIG                     = "/root/.kube/config"
    GOOGLE_APPLICATION_CREDENTIALS = try(env.GOOGLE_APPLICATION_CREDENTIALS, "")
    CLOUDFLARE_API_KEY             = try(env.CLOUDFLARE_API_KEY, "")
    CLOUDFLARE_EMAIL               = try(env.CLOUDFLARE_EMAIL, "")
    CLOUDFLARE_ACCOUNT_ID          = try(env.CLOUDFLARE_ACCOUNT_ID, "")
    AUTH0_APP_ID                   = try(env.AUTH0_APP_ID, "")
    AUTH0_DOMAIN                   = try(env.AUTH0_DOMAIN, "")
    AUTH0_MGMT_CLIENT_ID           = try(env.AUTH0_MGMT_CLIENT_ID, "")
    AUTH0_MGMT_CLIENT_SECRET       = try(env.AUTH0_MGMT_CLIENT_SECRET, "")
  }
  volumes = [
    "${workspace.path}:${workspace.path}",
    "/docker/certs:/docker/certs",
    "${try(env.WATCHMAN_SOCK, "")}:${try(env.WATCHMAN_SOCK, "")}",
    "/root/.vault-token:/root/.vault-token",
    "/buildkite/scratch/development/development_config:/root/.kube/config",
    "/etc/google:/etc/google"
  ]
  privileged = true
}
