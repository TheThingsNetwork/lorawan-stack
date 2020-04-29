---
title: "Installing the Command-line Interface"
description: ""
---

## Installing the Command-line interface

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

## Configuration

The command-line needs to be configured to connect to your deployment on `thethings.example.com`. You have multiple options to make the configuration file available to the CLI:

1. Environment: `export TTN_LW_CONFIG=/path/to/ttn-lw-cli.yml`
2. Command-line flag: `-c /path/to/ttn-lw-cli.yml`
3. Save as `.ttn-lw-cli.yml` in `$XDG_CONFIG_HOME`, your home directory, or the working directory.

> NOTE: When using the snap packages, `~/.ttn-lw-cli.yml` will fail with permission errors. Choose a different path.

### Generate configuration file

For a standard deployment on `thethings.example.com`, all you need is:

```bash
$ ttn-lw-cli use thethings.example.com [--fetch-ca] [--user] [--overwrite]
```

This will generate and save the required CLI config file. By default, the file is saved on the current directory, use the `--user` to save it under the user config directory.

If the deployment is using a CA that is not already trusted by your system, use the `--fetch-ca` flag to also connect to the server and retrieve the CA required for establishing secure communication.

> NOTE: If the file exists already, it is not overwritten and an error is printed instead. You can use `--overwrite` to overwrite the existing configuration file.

> NOTE: You can also use the `--grpc-port` and `--oauth-server-address` flags to override the default values for the gRPC port and the OAuth server address. These are not needed for standard deployments.

### Manually creating configuration file

Alternatively, you can create a file named `.ttn-lw-cli.yml` and paste the following contents:

```yaml
oauth-server-address: 'https://thethings.example.com/oauth'

identity-server-grpc-address: 'thethings.example.com:8884'
gateway-server-grpc-address: 'thethings.example.com:8884'
network-server-grpc-address: 'thethings.example.com:8884'
application-server-grpc-address: 'thethings.example.com:8884'
join-server-grpc-address: 'thethings.example.com:8884'
device-claiming-server-grpc-address: 'thethings.example.com:8884'
device-template-converter-grpc-address: 'thethings.example.com:8884'
qr-code-generator-grpc-address: 'thethings.example.com:8884'
```

For advanced options, see the [Configuration Reference]({{< ref "/reference/configuration/cli" >}}).