package "setup" {
  description = "The test package for the setup portion of the arksdk."

}

target "arktest" "example" {
  special = "testing"
}


target "arktest" "example2" {
  special = "testing2"
  depends_on = [
    "setup.arktest.example"
  ]
}

