---
layout: "functions"
page_title: "cidrsubnets - Functions - Configuration Language"
sidebar_current: "docs-funcs-ipnet-cidrsubnets"
description: |-
  The cidrsubnets function calculates a sequence of consecutive IP address
  ranges within a particular CIDR prefix.
---

# `cidrsubnets` Function

`cidrsubnets` calculates a sequence of consecutive IP address ranges within
a particular CIDR prefix.

```hcl
cidrsubnets(prefix, newbits...)
```

`prefix` must be given in CIDR notation, as defined in
[RFC 4632 section 3.1](https://tools.ietf.org/html/rfc4632#section-3.1).

The remaining arguments, indicated as `newbits` above, each specify the number
of additional network prefix bits for one returned address range. The return
value is therefore a list with one element per `newbits` argument, each
a string containing an address range in CIDR notation.

For more information on IP addressing concepts, see the documentation for the
related function [`cidrsubnet`](./cidrsubnet.md). `cidrsubnet` calculates
a single subnet address within a prefix while allowing you to specify its
subnet number, while `cidrsubnets` can calculate many at once, potentially of
different sizes, and assigns subnet numbers automatically.

When using this function to partition an address space as part of a network
address plan, you must not change any of the existing arguments once network
addresses have been assigned to real infrastructure, or else later address
assignments will be invalidated. However, you _can_ append new arguments to
existing calls safely, as long as there is sufficient address space available.

This function accepts both IPv6 and IPv4 prefixes, and the result always uses
the same addressing scheme as the given prefix.

## Examples

```
> cidrsubnets("10.1.0.0/16", 4, 4, 8, 4)
[
  "10.1.0.0/20",
  "10.1.16.0/20",
  "10.1.32.0/24",
  "10.1.48.0/20",
]

> cidrsubnets("fd00:fd12:3456:7890::/56", 16, 16, 16, 32)
[
  "fd00:fd12:3456:7800::/72",
  "fd00:fd12:3456:7800:100::/72",
  "fd00:fd12:3456:7800:200::/72",
  "fd00:fd12:3456:7800:300::/88",
]
```

## Related Functions

* [`cidrhost`](./cidrhost.md) calculates the IP address for a single host
  within a given network address prefix.
* [`cidrnetmask`](./cidrnetmask.md) converts an IPv4 network prefix in CIDR
  notation into netmask notation.
* [`cidrsubnet`](./cidrsubnet.md) calculates a single subnet address, allowing
  you to specify its network number.
