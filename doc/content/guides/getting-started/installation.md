---
title: "Installation"
description: ""
weight: 1
---

## Preparation

Since we're going to install {{% tts %}} using Docker and Docker Compose, follow the guides to [install Docker](https://docs.docker.com/install/#supported-platforms) and to [install Docker Compose](https://docs.docker.com/compose/install/#install-compose).

Most releases contain an example `docker-compose.yml` file. You can also find this file [in the Github repository of {{% tts %}}]({{% repo-file-url "raw" "docker-compose.yml" %}}). In this guide we'll use that example `docker-compose.yml` for our deployment.

## Command-line Interface (optional)

Although the web interface of {{% tts %}} (the Console) currently has support for all basic features of {{% tts %}}, for some actions, you need to use the command-line interface (CLI). The CLI allows you to manage all features of {{% tts %}}.

You can use the CLI on your local machine and on the server.

>Note: if you need help with any CLI command, use the `--help` flag to get a list of subcommands, flags and their description and aliases.

### Package managers (recommended)

#### macOS

```bash
$ brew install TheThingsNetwork/lorawan-stack/ttn-lw-stack
```

#### Linux

```bash
$ sudo snap install ttn-lw-stack
$ sudo snap alias ttn-lw-stack.ttn-lw-cli ttn-lw-cli
```

### Binaries

You can download [pre-built binaries](https://github.com/TheThingsNetwork/lorawan-stack/releases) for your operating system and processor architecture.
