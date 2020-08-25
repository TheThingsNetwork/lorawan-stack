---
title: "Adding Devices"
description: ""
aliases: [/getting-started/cli/create-end-device, /getting-started/console/create-end-device]
---

This section contains instructions for adding devices in {{% tts %}}.

<!--more-->

{{< tabs/container "Console" "CLI" >}}

{{< tabs/tab "Console" >}}

## Adding Devices using the Console

Go to **Devices** in the left menu and click on **+ Add Device** to reach the end device registration page. Fill the device ID, the LoRaWAN MAC and PHY versions and the frequency plan used by the device.

> Note: The PHY version represent the revision of the LoRAWAN Regional Parameters that the device expects, and must be correlated with the MAC version.

{{< figure src="device-creation-1.png" alt="Creating a new device" >}}

### Over-The-Air-Activation (OTAA) Device

After filling the fields in the **General Settings** section, scroll to the lower part of the device registration page and make sure that "Over The Air Activation (OTAA)" is selected. Fill the JoinEUI (AppEUI in LoRaWAN versions before 1.1), the DevEUI and AppKey. The NwkKey is only needed for LoRaWAN version 1.1 or later. All other fields on the page are optional. Press **Create Device** to create the device.

{{< figure src="device-creation-otaa.png" alt="Creating an OTAA device" >}}

You will now reach the device overview page for your device. The end device should now be able to join the private network.

>Note: If you do not have a `JoinEUI` or `AppEUI`, it is okay to use `0000000000000000`. Be sure to use the same `JoinEUI` in your device as you enter in {{% tts %}}.

{{< figure src="device-otaa-created.png" alt="OTAA device overview" >}}

### Activation By Personalization (ABP) Device

After filling the fields in the "General Settings" section, scroll to the lower part of the device registration page and make sure that "Activation By Personalization (ABP)" is selected. Fill the Device Address, the FNwkSIntKey (NwkSKey in LoRaWAN versions before 1.1) and the AppSKey. The other key fields are only needed for LoRaWAN version 1.1 or later. All other fields on the page are optional. Press **Create Device** to create the device.

{{< figure src="device-creation-abp.png" alt="Creating an ABP device" >}}

You will now reach the device overview page for your device. The end device should now be able to communicate with the private network.

{{< figure src="device-abp-created.png" alt="ABP device overview" >}}

{{< /tabs/tab >}}

{{< tabs/tab "CLI" >}}

## Adding Devices using the CLI

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

>Note: If you do not have a `JoinEUI` or `AppEUI`, it is okay to use `0000000000000000`. Be sure to use the same `JoinEUI` in your device as you enter in {{% tts %}}.

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

{{< /tabs/tab >}}

{{< /tabs/container >}}

Once a device has been added, get started with [Integrations]({{< ref "/integrations" >}}) to process and act on data.
