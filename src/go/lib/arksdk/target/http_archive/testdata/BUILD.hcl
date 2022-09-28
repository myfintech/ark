package "test" {
  description = "The test package for the local_file build target."
}

target "http_archive" "test" {
  url = "https://github.com/keplerproject/luafilesystem/archive/v1_6_2.tar.gz"
  sha256 = "7f2910e6c7fbc1d64d0a6535e6a514ed138051af13ee94bccdeb7a20146f18d9"
  decompress = true
}
