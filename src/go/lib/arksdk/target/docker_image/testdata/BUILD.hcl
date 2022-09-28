package "test" {
  description = "The test package for the docker_image build target."
}

locals {
  thing = "test"
  other_thing = 3
}

target "docker_image" "test" {
  repo = "gcr.io/[insert-google-project]/domain/${locals.thing}"
  dockerfile = "Dockerfile"
  build_args = {
    GIT_SHA = "whatever"
  }
  tags = [
    "things"
  ]
}
