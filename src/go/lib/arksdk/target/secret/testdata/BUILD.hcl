package "test" {
  description = ""
}

target "secret" "test_1" {
  optional = true
  secret_name = "file-test"
  files = ["./secretdata.json"]
}

target "secret" "test_2" {
  optional = true
  secret_name = "file-test"
  files = ["./missing.json"]
}

target "secret" "test_3" {
  optional = false
  secret_name = "file-test"
  files = ["./secretdata.json"]
}

target "secret" "test_4" {
  optional = false
  secret_name = "file-test"
  files = ["./missing.json"]
}

target "secret" "test_5" {
  optional = true
  secret_name = "env-test"
  environment = ["HOME"]
}

target "secret" "test_6" {
  optional = true
  secret_name = "env-test"
  environment = ["MISSING"]
}

target "secret" "test_7" {
  optional = false
  secret_name = "env-test"
  environment = ["HOME"]
}

target "secret" "test_8" {
  optional = false
  secret_name = "env-test"
  environment = ["MISSING"]
}

target "secret" "test_9" {
  optional = true
  secret_name = "fail-test-one"
}

target "secret" "test_10" {
  optional = true
  secret_name = "fail-test-two"
  files = ["./secretdata.json"]
  environment = ["HOME"]
}

target "secret" "test_11" {
  optional = true
  secret_name = "file-test"
  files = ["./secretdatadir"]
}
