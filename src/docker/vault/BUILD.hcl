package "dev_vault" {
  description = ""
}

target "build" "image" {
  repo = "gcr.io/[insert-google-project]/ark/dev/vault"
  dockerfile = file("${package.path}/Dockerfile")
  disable_entrypoint_injection = true
  source_files = [
    "./data",
    "./config.hcl",
    "./entrypoint.sh",
    "./unseal.json"
  ]
}
