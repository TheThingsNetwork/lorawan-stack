---
title: "Register Things"
description: ""
weight: 5
--- 

The stack count a total of three things; gateways, application, devices. Each is highly tunable, don't hesitate go to check the [References](../../references) to see all possible settings.
In this guide, we only use simple and standard configuration.

## <a name="creategtw">Creating a gateway</a>

First, list the available frequency plans:

```
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

This creates a gateway `gtw1` with user `admin` as collaborator, frequency plan `EU_863_870`, EUI `00800000A00009EF` and respecting duty-cycle limitations. You can now connect your gateway to the stack.

>Note: The CLI returns the created and updated entities by default in JSON. This can be useful in scripts.

## <a name="createapp">Creating an application</a>

Create the first application:

```bash
$ ttn-lw-cli applications create app1 --user-id admin
```

This creates an application `app1` with the `admin` user as collaborator.

Devices are created within applications.

## <a name="createdev">Creating a device</a>

First, list the available frequency plans and LoRaWAN versions:

```
$ ttn-lw-cli end-devices list-frequency-plans
$ ttn-lw-cli end-devices create --help
```

Then, to create an end device using over-the-air-activation (OTAA):

```bash
$ ttn-lw-cli end-devices create app1 dev1 \
  --dev-eui 0004A30B001C0530 \
  --app-eui 800000000000000C \
  --frequency-plan-id EU_863_870 \
  --root-keys.app-key.key 752BAEC23EAE7964AF27C325F4C23C9A \
  --lorawan-version 1.0.2 \
  --lorawan-phy-version 1.0.2-b
```

This will create a LoRaWAN 1.0.2 end device `dev1` in application `app1`. The end device should now be able to join the private network.

>Note: The `AppEUI` is returned as `join_eui` (V3 uses LoRaWAN 1.1 terminology).

>Hint: You can also pass `--with-root-keys` to have root keys generated.

It is also possible to register an ABP activated device using the `--abp` flag as follows:

```bash
$ ttn-lw-cli end-devices create app1 dev2 \
  --frequency-plan-id EU_863_870 \
  --lorawan-version 1.0.2 \
  --lorawan-phy-version 1.0.2-b \
  --abp \
  --session.dev-addr 00E4304D \
  --session.keys.app-s-key.key A0CAD5A30036DBE03096EB67CA975BAA \
  --session.keys.nwk-s-key.key B7F3E161BC9D4388E6C788A0C547F255
```

>Note: The `NwkSKey` is returned as `f_nwk_s_int_key` (V3 uses LoRaWAN 1.1 terminology).

>Hint: You can also pass `--with-session` to have a session generated.
