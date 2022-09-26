package "local_exec" {
  description = ""
}

target "build" "test_image" {
  repo = "gcr.io/managed-infrastructure/mantl/${package.name}-test"
  dockerfile = templatefile("${package.path}/Dockerfile_Test", {})

  source_files = [
    package.path,
    "${workspace.path}/WORKSPACE.hcl",
    "${workspace.path}/go.mod",
    "${workspace.path}/go.sum",
    "${workspace.path}/src/go/lib/log",
    "${workspace.path}/src/go/lib/arksdk/target/base",
    "${workspace.path}/src/go/lib/exec",
    "${workspace.path}/src/go/lib/hclutils",
    "${workspace.path}/src/go/lib/git/gitignore",
    "${workspace.path}/src/go/lib/container",
    "${workspace.path}/src/go/lib/dag",
    "${workspace.path}/src/go/lib/kube",
    "${workspace.path}/src/go/lib/fs/observer",
    "${workspace.path}/src/go/lib/utils",
    "${workspace.path}/src/jsonnet/lib",
    "${workspace.path}/src/go/lib/jsonnetutils",
    "${workspace.path}/src/go/lib/utils/cryptoutils",
    "${workspace.path}/src/go/lib/watchman",
    "${workspace.path}/src/go/lib/kube/portbinder",
    "${workspace.path}/src/go/lib/state_store",
    "${workspace.path}/src/go/lib/fs",
    "${workspace.path}/src/go/lib/utils/cloudutils",
    "${workspace.path}/src/go/lib/pattern",
    "${workspace.path}/src/go/lib/arksdk/kv",
    "${workspace.path}/src/go/lib/vault_tools/vault_test_harness",

  ]
}

//target "test" "local-exec" {
//  image = local_exec.build.test_image.url
//  command = ["go test -v ./..."]
//}
