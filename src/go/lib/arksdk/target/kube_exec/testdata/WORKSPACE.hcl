kubernetes {
  safe_contexts = [
    "local",
    "docker-desktop",
    "docker-for-desktop",
    "development_admin",
    "development_sre",
    "gke_[insert-google-project]_us-central1_development-us-central1"
  ]
}
