package "test" {
  description = "The test package for the target-base portion of the arksdk."
}

target "example" "foo" {
  not_a_raw_field = "testing"

  extra_fields = "testing"

  labels = ["test1", "test2", "test3"]

  source_files = [
    "./dont_touch_me.txt"
  ]

  exclude_patterns = [
    "*.go",
    "file-does-not-exist"
  ]

  depends_on = []
}
