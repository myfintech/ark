---
layout: "functions"
page_title: "filebase64sha256 - Functions - Configuration Language"
sidebar_current: "docs-funcs-crypto-filebase64sha256"
description: |-
  The filebase64sha256 function computes the SHA256 hash of the contents of
  a given file and encodes it with Base64.
---

# `filebase64sha256` Function

`filebase64sha256` is a variant of [`base64sha256`](./base64sha256.md)
that hashes the contents of a given file rather than a literal string.

This is similar to `base64sha256(file(filename))`, but
because [`file`](./file.md) accepts only UTF-8 text it cannot be used to
create hashes for binary files.
