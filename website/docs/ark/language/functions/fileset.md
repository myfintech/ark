---
layout: "functions"
page_title: "fileset - Functions - Configuration Language"
sidebar_current: "docs-funcs-file-file-set"
description: |-
  The fileset function enumerates a set of regular file names given a pattern.
---

# `fileset` Function

`fileset` enumerates a set of regular file names given a path and pattern.
The path is automatically removed from the resulting set of file names and any
result still containing path separators always returns forward slash (`/`) as
the path separator for cross-system compatibility.

```hcl
fileset(path, pattern)
```

Supported pattern matches:

- `*` - matches any sequence of non-separator characters
- `**` - matches any sequence of characters, including separator characters
- `?` - matches any single non-separator character
- `{alternative1,...}` - matches a sequence of characters if one of the comma-separated alternatives matches
- `[CLASS]` - matches any single non-separator character inside a class of characters (see below)
- `[^CLASS]` - matches any single non-separator character outside a class of characters (see below)

Character classes support the following:

- `[abc]` - matches any single character within the set
- `[a-z]` - matches any single character within the range

Functions are evaluated during configuration parsing rather than at apply time,
so this function can only be used with files that are already present on disk
before Ark takes any actions.

## Examples

```
> fileset(path.module, "files/*.txt")
[
  "files/hello.txt",
  "files/world.txt",
]

> fileset(path.module, "files/{hello,world}.txt")
[
  "files/hello.txt",
  "files/world.txt",
]

> fileset("${path.module}/files", "*")
[
  "hello.txt",
  "world.txt",
]

> fileset("${path.module}/files", "**")
[
  "hello.txt",
  "world.txt",
  "subdirectory/anotherfile.txt",
]
```
