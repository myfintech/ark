package "test" {
  description = "The test package for the jsonnet build target."
}

target "jsonnet" "test" {
  file = "test.jsonnet"
  variables = jsonencode({
    "color": "red",
    "terrain": 1,
    "arr": [1, null, "thing"]
  })
  library_dir = [
    "test_lib"
  ]
}