---
layout: "functions"
page_title: "matchkeys - Functions - Configuration Language"
sidebar_current: "docs-funcs-collection-matchkeys"
description: |-
  The matchkeys function takes a subset of elements from one list by matching
  corresponding indexes in another list.
---

# `matchkeys` Function

`matchkeys` constructs a new list by taking a subset of elements from one
list whose indexes match the corresponding indexes of values in another
list.

```hcl
matchkeys(valueslist, keyslist, searchset)
```

`matchkeys` identifies the indexes in `keyslist` that are equal to elements of
`searchset`, and then constructs a new list by taking those same indexes from
`valueslist`. Both `valueslist` and `keyslist` must be the same length.

The ordering of the values in `valueslist` is preserved in the result.

## Examples

```
> matchkeys(["i-123", "i-abc", "i-def"], ["us-west", "us-east", "us-east"], ["us-east"])
[
  "i-abc",
  "i-def",
]
```
