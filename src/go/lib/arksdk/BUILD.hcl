package "arksdk" {
  description = <<-EOT
  The arksdk serves as the foundation for the arkcli tool.
  All targets and their various attributes are defined within
  this SDK. Targets allow for various workflows to be executed.
  For example, there is a go_binary target that can build golang
  binaries, a docker_image target that can build docker images,
  and so on. The SDK tracks builds by hashing source files and
  writing a state file for the target. That state is verified at
  build time, and if the hash is unchanged, a target is not rebuilt.
EOT
}
