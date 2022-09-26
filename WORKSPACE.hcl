kubernetes {
  safe_contexts = [
    "local",
    "docker-desktop",
    "docker-for-desktop",
    "development_admin",
    "development_sre",
    "gke_managed-infrastructure_us-central1_development-us-central1"
  ]
}

file_system {
  ignore = [
    ".git/**",
    ".idea/**",
    "vendor/**",
    "**/node_modules/**",
    "**/*~",
    "**/*.swx",
    "**/*.swp"
  ]
}

artifacts {
  storage_base_url = "gs://ark-cache"
}

jsonnet {
  library = [
    "src/jsonnet"
  ]
}

vault {
  encryption_key = "sre-enc-key"
  address = "http://779c9729896f5cfc.mantl.sdm.network:65230"
}

plugin "vault-k8s-service-account" {
  image = "gcr.io/managed-infrastructure/ark/plugins/vault-k8s-sa:973721b27731109237e51e0dfe6891cdd2ce7dd0188a16b356b32b86c0fe6d32"
}

plugin "k8s-microservice" {
  image = "gcr.io/managed-infrastructure/ark/plugins/microservice:latest"
}

plugin "gcloud-emulator" {
  image = "gcr.io/managed-infrastructure/ark/plugins/gcloud-emulator:latest"
}

plugin "core-proxy" {
  image = "gcr.io/managed-infrastructure/ark/plugins/core-proxy:08a1113807725e925559a713b4c4042eb3be932226aa3cb4bdfc44cbdb6c486c"
}

plugin "actions-runner" {
  image = "gcr.io/managed-infrastructure/ark/plugins/gha-runner:latest"
}

plugin "datadog" {
  image = "gcr.io/managed-infrastructure/ark/plugins/datadog:latest"
}

plugin "vault" {
  image = "gcr.io/managed-infrastructure/ark/plugins/vault:latest"
}

plugin "postgres" {
  image = "gcr.io/managed-infrastructure/ark/plugins/postgres:latest"
}

plugin "kube-state" {
  image = "gcr.io/managed-infrastructure/ark/plugins/kube-state:08949ab"
}

plugin "terraform-cloud-agent" {
  image = "gcr.io/managed-infrastructure/ark/plugins/terraform-cloud-agent:5c2276e"
}

plugin "consul" {
  image = "gcr.io/managed-infrastructure/ark/plugins/consul:latest"
}

plugin "ns-reaper" {
  image = "gcr.io/managed-infrastructure/ark/plugins/ns-reaper:latest"
}
