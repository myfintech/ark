---
id: http_archive
title: HTTP Archive
sidebar_label: http_archive
---

# `http_archive`

Downloads an item from a remote URL.


## Example Usage

```hcl
target "http_archive" "confluent_operator" {
  url = "https://cdn.mantl.team/ark/external/confluent-operator/${locals.operator_name}.tar.gz"
  sha256 = "8031142b440bd41dd480f585954aeee90599f759ecf2d42acc2d69a1bfdf7d35"
}
```

## Attribute Reference

### Inputs

| Attribute | Required | Type | Explanation |
| --------- | :------: | ---- | ----------- |
| url | :heavy_check_mark: | `string` | The URL of the item to be downloaded. |
| sha256 | :heavy_check_mark: | `string` | The SHA256 of the item to be downloaded. The `http_archive` will use this attribute to compare against the downloaded artifact before moving the artifact to the `ark` storage volume. |
| decompress |  | `boolean` | If `decompress` is set, it will attempt to expand a compressed artifact into the proper `ark` storage volume location. |

### Outputs

| Attribute | Type | Explanation |
| --------- | ---- | ----------- |
| `url` | `string` | A direct output of the `url` input attribute if the URL needs to be referenced in another target. |
| `contents_path` | `string` | The filesystem path where a downloaded archive will be placed. |
