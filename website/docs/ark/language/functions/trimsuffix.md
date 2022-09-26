---
layout: "functions"
page_title: "trimsuffix - Functions - Configuration Language"
sidebar_current: "docs-funcs-string-trimsuffix"
description: |-
  The trimsuffix function removes the specified suffix from the end of a
  given string.
---

# `trimsuffix` Function

`trimsuffix` removes the specified suffix from the end of the given string.

## Examples

```
> trimsuffix("helloworld", "world")
hello
```

## Related Functions

* [`trim`](./trim.md) removes characters at the start and end of a string.
* [`trimprefix`](./trimprefix.md) removes a word from the start of a string.
* [`trimspace`](./trimspace.md) removes all types of whitespace from
  both the start and the end of a string.
