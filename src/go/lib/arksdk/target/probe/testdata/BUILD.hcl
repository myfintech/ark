package "test" {
  description = "The test package for the probe build target."
}

target "probe" "tcp" {
  delay = "2s"
  timeout = "5s"
  address = "tcp://0.0.0.0:31200"
  max_retries = 10
}

target "probe" "http" {
  delay = "2s"
  timeout = "5s"
  address = "http://0.0.0.0:31200"
  expected_status = 404
  max_retries = 10
}