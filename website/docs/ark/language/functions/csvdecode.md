---
layout: "functions"
page_title: "csvdecode - Functions - Configuration Language"
sidebar_current: "docs-funcs-encoding-csvdecode"
description: |-
  The csvdecode function decodes CSV data into a list of maps.
---

# `csvdecode` Function

`csvdecode` decodes a string containing CSV-formatted data and produces a
list of maps representing that data.

CSV is _Comma-separated Values_, an encoding format for tabular data. There
are many variants of CSV, but this function implements the format defined
in [RFC 4180](https://tools.ietf.org/html/rfc4180).

The first line of the CSV data is interpreted as a "header" row: the values
given are used as the keys in the resulting maps. Each subsequent line becomes
a single map in the resulting list, matching the keys from the header row
with the given values by index. All lines in the file must contain the same
number of fields, or this function will produce an error.

## Examples

```
> csvdecode("a,b,c\n1,2,3\n4,5,6")
[
  {
    "a" = "1"
    "b" = "2"
    "c" = "3"
  },
  {
    "a" = "4"
    "b" = "5"
    "c" = "6"
  }
]
```
