package "test" {
  description = "The test package for the local_file build target."
}

locals {
  thing = "test"
  other_thing = 3
  plugin_pass_test = {
    "array": [
      1,
      2,
      3
    ],
    "boolean": true,
    "null": null,
    "number": 123,
    "object": {
      "a": "b",
      "c": "d",
      "e": "f"
    },
    "string": "Hello World"
  }
}

target "local_file" "test1" {
  filename = "thing.txt"
  content = "whatever"
}

target "local_file" "test2" {
  filename = "/tmp/thing2.txt"
  content = "whatever"
}

target "local_file" "test3" {
  filename = "thing3.txt"
  content = test(locals.plugin_pass_test)
}
