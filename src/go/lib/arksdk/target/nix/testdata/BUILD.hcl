package "test" {
  description = ""
}

target "nix" "test_cached" {
  packages = [
    "nixpkgs.watchman",
    "nixpkgs.k9s",
  ]
}
