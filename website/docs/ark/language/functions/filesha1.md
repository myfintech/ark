---
layout: "functions"
page_title: "filesha1 - Functions - Configuration Language"
sidebar_current: "docs-funcs-crypto-filesha1"
description: |-
  The filesha1 function computes the SHA1 hash of the contents of
  a given file and encodes it as hex.
---

# `filesha1` Function

`filesha1` is a variant of [`sha1`](./sha1.md)
that hashes the contents of a given file rather than a literal string.

This is similar to `sha1(file(filename))`, but
because [`file`](./file.md) accepts only UTF-8 text it cannot be used to
create hashes for binary files.
