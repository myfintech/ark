package "test" {
  description = "test"
}

module "mod" {
  source = "${workspace.path}/MODULE.hcl"
  repo_base = "test"
}
