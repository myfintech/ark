---
layout: "functions"
page_title: "artifacts - Functions - Configuration Language"
sidebar_current: "docs-funcs-file-artifacts"
description: |-
  The artifacts function returns a filesystem path based on a target's artifacts directory.
---

# `artifacts` Function

`artifacts` takes a string and converts it
to an absolute path based on a target's artifacts directory. 

## Examples

```
> artifacts("bin")
/ark/artifacts/<target>/<hash>/bin
> artifacts("")
/ark/artifacts/<target>/<hash>
```
