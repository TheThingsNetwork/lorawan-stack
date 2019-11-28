---
title: "Command-line Interface"
description: ""
weight: 10
---

## Configuration

If you have not configured the CLI yet, see the [**Configuration** step]({{< relref "../configuration.md#command-line-interface" >}}).

## Login

The CLI needs to be logged on in order to create gateways, applications, devices and API keys. With The Things Stack running in one terminal session, login with the following command:

```bash
$ ttn-lw-cli login
```

This will open the OAuth login page where you can login with your credentials. Once you logged in in the browser, return to the terminal session to proceed.

If you run this command on a remote machine, pass `--callback=false` to get a link to login on your local machine.

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

This creates a gateway `gtw1` with user `admin` as collaborator, frequency plan `EU_863_870`, EUI `00800000A00009EF` and respecting duty-cycle limitations. You can now connect your gateway to The Things Stack.

>Note: The CLI returns the created and updated entities by default in JSON. This can be useful in scripts.

### Create Gateway API Key

Some gateways require an API Key with Link Gateway Rights to be able to connect to The Things Stack.

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

>Note: The `NwkSKey` is returned as `f_nwk_s_int_key` (The Things Stack uses LoRaWAN 1.1 terminology).

>Hint: You can also pass `--with-session` to have a session generated.

## Working With Data

With your The Things Stack setup, a gateway connected and a device registered on your network, it's time to start working with data.

Learn how to work with the [builtin MQTT server]({{< relref "../mqtt" >}}) and [HTTP webhooks]({{< relref "../webhooks" >}}).
