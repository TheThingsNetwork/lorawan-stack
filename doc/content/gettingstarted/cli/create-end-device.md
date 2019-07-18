---
title: "Creating an end device"
description: ""
weight: 5
---

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

It is also possible to create a multicast device (an ABP device which can not send uplinks and shares the security session with other devices) using the `--multicast` flag as follows:

```bash
$ ttn-lw-cli end-devices create app1 dev3 \
  --frequency-plan-id EU_863_870 \
  --lorawan-version 1.0.2 \
  --lorawan-phy-version 1.0.2-b \
  --abp \
  --session.dev-addr 00E4304D \
  --session.keys.app-s-key.key A0CAD5A30036DBE03096EB67CA975BAA \
  --session.keys.nwk-s-key.key B7F3E161BC9D4388E6C788A0C547F255 \
  --multicast
```

>Note: The `--multicast` flag can be set only during device creation, and as such can not be turned on or off later.
