---
id: quick-start
title: Quick Start
sidebar_label: Quick Start
---

# Pre-Install

- Download docker community edition for your OS

## (optional) high performance file system operations 
- Facebook watchman is a high performance file system observer.
- This is highly beneficial when projects have hundreds of thousands of files.
- If it is available in your `$PATH` ark will use it to query the file system and subscribe to change notifications.

- Install [homebrew](https://brew.sh/) on macos
```bash
brew install watchman
```

# Installing Ark

**NOTE**
To keep ark up to date a version check executes after every command.

`Hover over the code block to copy the command to your clipboard.`

```bash
curl -L https://cdn.mantl.team/assets/arkcli/install.sh | sh
```