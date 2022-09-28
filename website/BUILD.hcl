package "website" {
  description = "website"
}

locals {
  workspace = "/opt/workspace"
  repo      = "gcr.io/[insert-google-project]/engineering/website"
}

target "build" "dev_image" {
  repo = locals.repo
  tags = ["development"]

  // language=docker
  dockerfile = <<-DOCKER
  FROM node:10.21.0-alpine as build
  RUN apk add git
  # Copy package locks for efficnent build
  COPY ./website/package.json ${locals.workspace}/website/package.json
  COPY ./website/yarn.lock ${locals.workspace}/website/yarn.lock
  RUN yarn --cwd ${locals.workspace}/website

  # After building node_modules copy the context and run the build
  COPY / ${locals.workspace}
  RUN yarn --cwd ${locals.workspace}/website build
  DOCKER

  source_files = [
    "./yarn.lock",
    "./package.json",
    "./babel.config.js",
    "./docusaurus.config.js",
    "./sidebars.js",
    "./README.md",
    "./blog",
    "./docs",
    "./src",
    "./static"
  ]
}

target "build" "release_image" {
  repo = locals.repo
  tags = ["release"]

  // language=docker
  dockerfile = <<-DOCKER
  # After build copy only the artifacts to the root of a scratch image
  FROM ${website.build.dev_image.url} as build
  FROM scratch as release
  COPY --from=build ${locals.workspace}/website/build /
  DOCKER

  output = artifact("")
}

target "docker_exec" "publish" {
  image             = "google/cloud-sdk:297.0.1-alpine"
  working_directory = locals.workspace
  volumes = [
    "${env.HOME}/.config/gcloud:/root/.config/gcloud",
    "${website.build.release_image.artifacts.path}:${locals.workspace}"
  ]

  command = [
    "sh",
    "-c",
    // language=bash
    <<-INSTALL
    gsutil -m -h "Cache-Control:no-cache,max-age=0" \
      rsync -d -r ./ gs://[INSERT_DOMAIN]
    INSTALL
  ]
}

target "docker_exec" "serve" {
  image             = website.build.dev_image.url
  working_directory = "${locals.workspace}/website"
  command = [
    "sh",
    "-c",
    <<-CMD
    echo "Host URL: http://localhost:3002" && \
    echo "This may take a moment to compile (be patient)" && \
    yarn start --host 0.0.0.0
  CMD
  ]
  ports = ["3002:3000"]
}
