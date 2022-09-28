package "test" {
  description = "The test package for the docker_image build target."
}

locals {
  thing = "test"
  other_thing = 3
  values = [
    {
      src = "./"
      dest = "./"
    }]
}

target "build" "test1" {
  repo = "gcr.io/[insert-google-project]/${locals.thing}"

  dockerfile = templatefile("${package.path}/Dockerfile", {
    from = "alpine"
    sources = [for file in locals.values: file]
    commands = [
      "apk --no-cache add jq",
      "--mount=type=secret,id=test cat /run/secrets/test | jq '.test'"
    ]
  })

  dump_context = true

  secrets = {
    "test": "sre/golang-tests/dummy-secret"
  }

  tags = [
    "things"
  ]

  source_files = [
    "${package.path}/BUILD.hcl"
  ]

  cache_inline = true
}

target "build" "test2" {
  repo = "gcr.io/[insert-google-project]/${locals.thing}"

  dockerfile = templatefile("${package.path}/Dockerfile", {
    from = "alpine"
    sources = [for file in locals.values: file]
    commands = [
      "apk update",
    ]
  })

 tags = [
    "things"
  ]

  source_files = [
    "${package.path}/BUILD.hcl"
  ]

  cache_inline = true
}
