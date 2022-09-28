package "test" {
  description = "The test package for the k8s deployment build target."
}

locals {
  command = "yarn"
  patterns = ["package.json", "*lock*"]
}

target "deploy" "test1" {
  manifest = file("./manifest.yaml")
  port_forward = ["8080:80", "8443:443", "3000:3000"]

  live_sync_enabled = false
  live_sync_restart_mode = "delegated"
  live_sync_on_actions = [{
    command = [locals.command]
    work_dir = "./"
    patterns = locals.patterns
  }]
//
//  env = [{
//    name = "TEST"
//    value = "TEST"
//  }]
}
