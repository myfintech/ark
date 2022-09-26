package "ark-entrypoint-lib" {
  description = ""
}

target "exec" "protoc" {
  command = [
    "bash",
    "-c",
    "protoc -I ${workspace.path}/src/proto/ark entrypoint.proto --go_out=plugins=grpc:${package.path}"
  ]
}