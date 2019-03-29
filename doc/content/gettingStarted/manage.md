---
title: "Authenticate & Manage"
description: ""
weight: 3
draft: false
--- 

## <a name="login">Login using the CLI</a>

The CLI needs to be logged on in order to create gateways, applications, devices and API keys. With the stack running in one terminal session, login with the following command:

```bash
$ ttn-lw-cli login
```

A link will be provided to the OAuth login page where you can login using the credentials from the step ahead. Once you logged in in the browser, return to the terminal session to proceed.

## <a name="registergtw">Registering a gateway</a>

By default, the stack allows unregistered gateways to connect, but without providing a default band. As such, it is highly recommended that each gateway is registered:

```bash
$ ttn-lw-cli gateway create gtw1 --user-id admin --frequency-plan-id EU_863_870 --gateway-eui 00800000A00009EF --enforce-duty-cycle
```

This creates a gateway `gtw1` with the frequency plan `EU_863_870` and EUI `00800000A00009EF` that respects duty-cycle limitations. You can now connect your gateway to the stack.

The frequency plan is fetched automatically from the [configured source](#frequencyplans).

>Note: if you need help with any command in `ttn-lw-cli`, use the `--help` flag to get a list of subcommands, flags and their description and aliases.

## <a name="registerapp">Registering an application</a>

In order to register a device, an application must be created first:

```bash
$ ttn-lw-cli app create app1 --user-id admin
```

This creates an application `app1` for the user `admin`.

## <a name="registerdev">Registering a device</a>

You can now register an OTAA activated device to be used with the stack as follows:

```bash
$ ttn-lw-cli end-devices create app1 dev1 --dev-eui 0004A30B001C0530 --join-eui 800000000000000C --frequency-plan-id EU_863_870 --root-keys.app-key.key 752BAEC23EAE7964AF27C325F4C23C9A --lorawan-phy-version 1.0.2-b --lorawan-version 1.0.2
```

This will create an LoRaWAN 1.0.2 end device `dev1` with DevEUI `0004A30B001C0530`, AppEUI `800000000000000C` and AppKey `752BAEC23EAE7964AF27C325F4C23C9A`. After configuring the credentials in the end device, you should be able to join the private network.

It is also possible to register an ABP activated device using the `--abp` flag as follows:

```bash
$ ttn-lw-cli end-devices create app1 dev1 --frequency-plan-id EU_863_870 --lorawan-phy-version 1.0.2-b --lorawan-version 1.0.2 --abp --session.dev-addr 00E4304D --session.keys.app-s-key.key A0CAD5A30036DBE03096EB67CA975BAA --session.keys.f_nwk_s_int_key.key B7F3E161BC9D4388E6C788A0C547F255
```

This will create an LoRaWAN 1.0.2 end device `dev1` with DevAddr `00E4304D`, AppSKey `A0CAD5A30036DBE03096EB67CA975BAA` and NwkSKey `B7F3E161BC9D4388E6C788A0C547F255`.

## <a name="linkappserver">Linking the application</a>

In order to send uplinks and receive downlinks from your device, you must first link the application server to the network server. In order to do this, first create an API key for the application server:

```bash
$ ttn-lw-cli app api-keys create --application-id app1 --right-application-link
```

The CLI will return an API key such as `NNSXS.VEEBURF3KR77ZR...`. This API key has only link rights and can therefore only be used during the linking process. 

You can now link the application server to the network server:

```bash
$ ttn-lw-cli app link set app1 --api-key NNSXS.VEEBURF3KR77ZR..
```

Your application is now linked, and can use the built-in MQTT broker and webhooks support to receive uplink traffic and send downlink traffic.



