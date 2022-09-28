kubernetes {
  safe_contexts = [
    "local",
    "docker-desktop",
    "docker-for-desktop",
  ]
}

file_system {
  ignore = [
    ".git/**",
    ".idea/**",
    "vendor/**",
    "**/node_modules/**",
    "**/*~",
    "**/*.swx",
    "**/*.swp"
  ]
}

artifacts {
  storage_base_url = "gs://ark-cache"
}

jsonnet {
  library = [
    "src/jsonnet"
  ]
}
