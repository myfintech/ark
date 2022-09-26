ui = true
disable_mlock = true

storage "file" {
  path = "/mnt/vault/data"
}


listener "tcp" {
  address = "0.0.0.0:8200"
  tls_disable = true
}
