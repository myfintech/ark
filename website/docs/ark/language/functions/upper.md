---
layout: "functions"
page_title: "upper - Functions - Configuration Language"
sidebar_current: "docs-funcs-string-upper"
description: |-
  The upper function converts all cased letters in the given string to uppercase.
---

# `upper` Function

`upper` converts all cased letters in the given string to uppercase.

## Examples

```
> upper("hello")
HELLO
> upper("алло!")
АЛЛО!
```

This function uses Unicode's definition of letters and of upper and lowercase.

## Related Functions

* [`lower`](./lower.md) converts letters in a string to _lowercase_.
* [`title`](./title.md) converts the first letter of each word in a string to uppercase.
