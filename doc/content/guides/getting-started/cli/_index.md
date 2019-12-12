---
title: "Command-line Interface"
description: ""
weight: 10
---

Although the web interface of {{% tts %}} (the Console) currently has support for all basic features of {{% tts %}}, for some actions, you need to use the command-line interface (CLI). The CLI allows you to manage all features of {{% tts %}}.
You can use the CLI on your local machine and on the server.

>Note: if you need help with any CLI command, use the `--help` flag to get a list of subcommands, flags and their description and aliases.

## Installation

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

## Login

The CLI needs to be logged on in order to create gateways, applications, devices and API keys. With {{% tts %}} running in one terminal session, login with the following command:

```bash
$ ttn-lw-cli login
```

This will open a browser window with the OAuth login page where you can login with your credentials. This is also where you can create a new account if you do not already have one.

> During the login procedure, the CLI starts a webserver on `localhost` in order to receive the OAuth callback after login. If you are running the CLI on a machine that is not `localhost`, you can pass the `--callback=false` flag. This will allow you to perform part of the OAuth flow on a different machine, and copy-paste a code back into the CLI.

## Create Gateway

First, list the available frequency plans:

```bash
$ ttn-lw-cli gateways list-frequency-plans
```

Then, create the first gateway with the chosen frequency plan:

```bash
$ ttn-lw-cli gateways create gtw1 \
  --user-id admin \
  --frequency-plan-id EU_863_870 \
  --gateway-eui 00800000A00009EF \
  --enforce-duty-cycle
```

This creates a gateway `gtw1` with user `admin` as collaborator, frequency plan `EU_863_870`, EUI `00800000A00009EF` and respecting duty-cycle limitations. You can now connect your gateway to {{% tts %}}.

>Note: The CLI returns the created and updated entities by default in JSON. This can be useful in scripts.

### Create Gateway API Key

Some gateways require an API Key with Link Gateway Rights to be able to connect to {{% tts %}}.

Create an API key for the gateway:

```bash
$ ttn-lw-cli gateways api-keys create \
  --name link \
  --gateway-id gtw1 \
  --right-gateway-link
```

The CLI will return an API key such as `NNSXS.VEEBURF3KR77ZR...`. This API key has only link rights and can therefore only be used for linking this gateway. Make sure to copy the key and save it in a safe place. You will not be able to see this key again in the future, and if you lose it, you can create a new one to replace it in the gateway configuration.

## Create Application

Create the first application:

```bash
$ ttn-lw-cli applications create app1 --user-id admin
```

This creates an application `app1` with the `admin` user as collaborator.

Devices are created within applications.

### Link Application

In order to send uplinks and receive downlinks from your device, you must link the Application Server to the Network Server. In order to do this, create an API key for the Application Server:

```bash
$ ttn-lw-cli applications api-keys create \
  --name link \
  --application-id app1 \
  --right-application-link
```

The CLI will return an API key such as `NNSXS.VEEBURF3KR77ZR...`. This API key has only link rights and can therefore only be used for linking this application. Make sure to copy the key and save it in a safe place. You will not be able to see this key again in the future, and if you lose it, you can create a new one to replace it in the gateway configuration.

You can now link the Application Server to the Network Server:

```bash
$ ttn-lw-cli applications link set app1 --api-key NNSXS.VEEBURF3KR77ZR..
```

Your application is now linked. You can now use the builtin MQTT server and webhooks to receive uplink traffic and send downlink traffic.

## Create End Device

First, list the available frequency plans and LoRaWAN versions:

```bash
$ ttn-lw-cli end-devices list-frequency-plans
$ ttn-lw-cli end-devices create --help
```

### Over-The-Air-Activation (OTAA) Device

To create an end device using over-the-air-activation (OTAA):

```bash
$ ttn-lw-cli end-devices create app1 dev1 \
  --dev-eui 0004A30B001C0530 \
  --app-eui 800000000000000C \
  --frequency-plan-id EU_863_870 \
  --root-keys.app-key.key 752BAEC23EAE7964AF27C325F4C23C9A \
  --lorawan-version 1.0.3 \
  --lorawan-phy-version 1.0.3-a
```

This will create a LoRaWAN 1.0.3 end device `dev1` in application `app1` with the `EU_863_870` frequency plan.

The end device should now be able to join the private network.

>Note: The `AppEUI` is returned as `join_eui` (V3 uses LoRaWAN 1.1 terminology).

>Hint: You can also pass `--with-root-keys` to have root keys generated.

### Activation By Personalization (ABP) Device

It is also possible to register an ABP activated device using the `--abp` flag as follows:

```bash
$ ttn-lw-cli end-devices create app1 dev2 \
  --frequency-plan-id EU_863_870 \
  --lorawan-version 1.0.3 \
  --lorawan-phy-version 1.0.3-a \
  --abp \
  --session.dev-addr 00E4304D \
  --session.keys.app-s-key.key A0CAD5A30036DBE03096EB67CA975BAA \
  --session.keys.nwk-s-key.key B7F3E161BC9D4388E6C788A0C547F255
```

>Note: The `NwkSKey` is returned as `f_nwk_s_int_key` ({{% tts %}} uses LoRaWAN 1.1 terminology).

>Hint: You can also pass `--with-session` to have a session generated.
