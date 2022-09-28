target "build" "test" {
  repo = module.vars.repo_base
  dockerfile = "from node"
}
