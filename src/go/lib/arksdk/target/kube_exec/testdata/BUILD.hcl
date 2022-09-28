package "test" {
  description = ""
}

target "kube_exec" "test" {
  resource_type = "ds"
  resource_name = "nginx-ingress"
  container_name = ""
  command = ["sh", "-c", "ls -lha"]
  get_pod_timeout = "10s"
}

target "kube_exec" "fail_test" {
  resource_type = "ds"
  resource_name = "nginx-ingress"
  container_name = ""
  command = ["sh", "-c", "ls -lha | grep -o 'fail'"]
  get_pod_timeout = "10s"
}