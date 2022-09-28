package "watchman" {
  description = <<-DESC
  Facebook watchman
  DESC
}

target "build" "image" {
  repo       = "gcr.io/[insert-google-project]/facebook/watchman"
  dockerfile = file("${package.path}/Dockerfile")
  source_files = [
    "./Dockerfile"
  ]
}
