---
layout: "functions"
page_title: "trim - Functions - Configuration Language"
sidebar_current: "docs-funcs-string-trim"
description: |-
  The trim function removes the specified characters from the start and end of
  a given string.
---

# `trim` Function

`trim` removes the specified characters from the start and end of the given
string.

## Examples

```
> trim("?!hello?!", "!?")
hello
```

## Related Functions

* [`trimprefix`](./trimprefix.md) removes a word from the start of a string.
* [`trimsuffix`](./trimsuffix.md) removes a word from the end of a string.
* [`trimspace`](./trimspace.md) removes all types of whitespace from
  both the start and the end of a string.
