package "test" {
  description = "The test package for the jsonnet build target."
}

target "jsonnet" "test" {
  yaml = true
  files = [
    "${workspace.path}/test.jsonnet"]
  output_dir = workspace.path
  variables = <<-DATA
  {
    "color": "red",
    "terrain": 1,
    "arr": [1, null, "thing"]
  }
  DATA
  library_dir = [
    "test_lib"
  ]
}