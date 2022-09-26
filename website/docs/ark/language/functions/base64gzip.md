---
layout: "functions"
page_title: "base64gzip - Functions - Configuration Language"
sidebar_current: "docs-funcs-encoding-base64gzip"
description: |-
  The base64encode function compresses the given string with gzip and then
  encodes the result in Base64.
---

# `base64gzip` Function

`base64gzip` compresses a string with gzip and then encodes the result in
Base64 encoding.

Ark uses the "standard" Base64 alphabet as defined in
[RFC 4648 section 4](https://tools.ietf.org/html/rfc4648#section-4).

Strings in HCL are sequences of unicode characters rather
than bytes, so this function will first encode the characters from the string
as UTF-8, then apply gzip compression, and then finally apply Base64 encoding.

While we do not recommend manipulating large, raw binary data in the HCL,
this function can be used to compress reasonably sized text strings
generated within HCL. For example, the result of this
function can be used to create a compressed object in Amazon S3 as part of
an S3 website.

## Related Functions

* [`base64encode`](./base64encode.md) applies Base64 encoding _without_
  gzip compression.
* [`filebase64`](./filebase64.md) reads a file from the local filesystem
  and returns its raw bytes with Base64 encoding.