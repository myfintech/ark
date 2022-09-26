package "watchman" {
  description = <<-DESC
  Facebook watchman
  DESC
}

target "build" "image" {
  repo       = "gcr.io/managed-infrastructure/facebook/watchman"
  dockerfile = file("${package.path}/Dockerfile")
  source_files = [
    "./Dockerfile"
  ]
}