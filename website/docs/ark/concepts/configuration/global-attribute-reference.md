---
id: global-attribute-reference
title: Global Target Attributes
sidebar_label: Global Attributes
---

Targets in `ark` like `build`, `docker_exec`, and `jsonnet_file` are distinguished by the required and optional attributes you supply in the `BUILD.hcl` file. For example, the `jsonnet_file` target requires a defined `file` attribute. If that attribute is defined in a `build` target, processing the target will fail because `file` isn't a recognized attribute for the `build` target.  

There are, however, several attributes that are available for all targets. Whether it makes sense to use them depends on how the target is being implemented.

# Global Attribute Reference

## Inputs

The following global attributes are optionally available for all targets:

| Attribute | Type | Explanation |
| --------- | ---- | ----------- |
| `description` | `string` | A blurb about what the target is meant to accomplish. Including descriptions can make understanding how a target is working easier. |
| `depends_on` | `array of strings` | If an implied dependency is not able to be derived from the target's definition, including a `depends_on` array allows a user to express that one or more targets should be built ahead of the target using the `depends_on`. |
| `declared_artifacts` | `array of strings` | A list of artifacts expected to be produced on a successful run of the target. |
| `source_files` | `array of strings` | A list of file path patterns that should be considered as a target's source material. This attribute is very useful for defining `build` targets. |\
| `include_patterns` | `array of strings` | A list of patterns that should be evaluated when determining if a file is a source file. For example, if `include_patterns` included `*.go`, and `source_files` was defined as `${package.path}/*`, only golang files within the package's path would be included. |
| `exclude_patterns` | `array of strings` | A list of patterns that should be evaluated when determining if a file should be excluded from source files. This attribute works the opposite way as `include_patterns`. |

## Outputs

There is a wealth of output attributes available for reference that come from several places. There are workspace output attributes, global target output attributes, and what are called HCL evaluation context output attributes.

### Workspace Output Attributes

#### Usage

`workspace.<output attribute>`

#### Workspace Output Attribute Reference

| Attribute | Type | Explanation |
| --------- | ---- | ----------- |
| `path` | `string` | The filesystem path to the workspace root. |
| `config.kubernetes.safe_contexts` | `array of strings` | A list of safe Kubernetes contexts as defined in the `WORKSPACE.hcl` file. |

### Global Target Output Attributes

#### Usage

`self.<output attribute>`

#### Global Target Output Attribute Reference

| Attribute | Type | Explanation |
| --------- | ---- | ----------- |
| `name` | `string` | The value of the target's 'name' label. The format for identifying target labels is as follows: `target "<type>" "<name>" {}` |
| `type` | `string` | The value of the target's 'type' label. The format for identifying target labels is as follows: `target "<type>" "<name>" {}` |
| `hash` | `string` | The current hash of the target. |
| `short_hash` | `string` | A truncated version of the `hash` attribute. |
| `dir` | `string` | The absolute filesystem path to get to the root of the current target's `BUILD.hcl` file. |
| `dir_rel` | `string` | The relative filesystem path to get to the root of the current target's `BUILD.hcl` file. The `dir_rel` attribue is relative to the workspace root. |
| `path` | `string` | The absolute filesystem path to get to the root of the current target's `BUILD.hcl` file. |
| `address` | `string` | The `ark` address of a target, expressed as `<package-name>.<target type>.<target name>`. |
| `artifacts` | `map of strings` | A reference object to where declared artifacts are stored for a given target. The path can be accessed with the following usage: `self.artifacts.<declared artifact name>`. |
| `artifacts.path` | `string` | The filesystem path to a targets default artifacts directory. |
| `source_files` | `array of strings` | The evaluated list of source files declared for a target. |

### HCL Evaluation Context Outputs

HCL Evaluation Context outputs aggregate all available outputs into a single context for a target.

| Attribute | Type | Usage | Explanation |
| --------- | ---- | ----- | ----------- |
| `locals` | `map of strings` | `locals.<local variable name>` | A block of local variables can be declared in a package, and those variables can be referenced using the `locals` attribute. |
| `workspace` | `map of strings` | `workspace.<workspace output attribute>` | See [workspace output attributes](#workspace-output-attributes) for information on specific workspace output attributes. |
| `runtime` | `map of strings` | `runtime.os` OR `runtime.arch` | Golang-specific output for obtaining the OS and system architecture of the current runtime. |
| `env` | `map of strings` | `env.<ENV VAR>` | Queries the session environment for a specific environment variable's value. |
| `self` | `map of strings` | `self.<global target output attribute>` | See [global target output attributes](#global-target-output-attributes) for information on specific global target output attributes. |
| `package` | `map of strings` | `package` OR `package.name` OR `package.version` OR `package.path` | Returns a map of the current package's name, version, and path OR each of those keys individually if dot notation is used to access a specific key. |
| `cli.args` | `string` | `cli.args` | Any arguments passed into the cli after a `--` indicator are available to any target using the `cli.args` attribute. This allows a user to create a generic target that takes in specific arguments when the target is run. |

# Additional Information Regarding Pattern Matching

`ark` uses the [glob][glob_lib] library to perform pattern matching.

Here are some examples: 

| Pattern | Fixture | Match |
| ------- | ------- | ----- |
| `[a-z][!a-x]*cat*[h][!b]*eyes*` | `my cat has very bright eyes` | `true` |
| `[a-z][!a-x]*cat*[h][!b]*eyes*` | `my dog has very bright eyes` | `false` |
| `https://*.google.*` | `https://account.google.com` | `true` |
| `https://*.google.*` | `https://google.com` | `false` |
| `{https://*.google.*,*yandex.*,*yahoo.*,*mail.ru}` | `http://yahoo.com` | `true` |
| `{https://*.google.*,*yandex.*,*yahoo.*,*mail.ru}` | `http://google.com` | `false` |
| `{https://*gobwas.com,http://exclude.gobwas.com}` | `https://safe.gobwas.com` | `true` |
| `{https://*gobwas.com,http://exclude.gobwas.com}` | `http://safe.gobwas.com` | `false` |
| `abc*` | `abcdef` | `true` |
| `abc*` | `af` | `false` |
| `*def` | `abcdef` | `true` |
| `*def` | `af` | `false` |
| `ab*ef` | `abcdef` | `true` |
| `ab*ef` | `af` | `false` |

[glob_lib]: github.com/gobwas/glob