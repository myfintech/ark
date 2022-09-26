package "event_sink_server" {
  description = ""
}

// TODO: make this a docker exec so there's not a local protoc dependency
target "exec" "generate_from_proto" {
  command = [
    "bash",
    "-c",
    "protoc -I ${workspace.path}/src/proto/ark ${package.name}.proto --go_out=plugins=grpc:${package.path}"
  ]
}
