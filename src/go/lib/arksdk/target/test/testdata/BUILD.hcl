package "test" {
  description = "The test package for the test target."
}

target "test" "test1" {
  image = "bash:latest"
  command = ["ls -lah | grep -o 'lib' && [[ $PWD == '/usr' ]] && [[ $$TEST == foo ]] && echo yes || echo fail"]
  timeout = 5
  working_directory = "/usr"
  environment = {
    TEST = "foo"
  }
}

target "test" "test2" {
  image = "bash:latest"
  command = ["ls -lah | grep -o 'fail'"]
  timeout = 5
}
