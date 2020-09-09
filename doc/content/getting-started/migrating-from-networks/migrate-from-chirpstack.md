---
title: "Migrate from ChirpStack"
description: ""
weight: 3
---

This section contains instructions on how to migrate end devices from ChirpStack to {{% tts %}}.

<!--more-->

End devices and applications can easily be migrated from ChirpStack to {{% tts %}} with the `ttn-lw-migrate` tool. This tool is used for exporting end devices and applications to a [JSON file]({{< ref "getting-started/migrating-from-networks/device-json.md" >}}) containing their description. This file can later be imported in {{% tts %}} as described in the [Import End Devices in The Things Stack]({{< ref "getting-started/migrating-from-networks/import-devices.md" >}}) section.

First, configure the environment with the following variables modified according to your setup:

```bash
$ export CHIRPSTACK_API_URL="localhost:8080"    # ChirpStack Application Server URL
$ export CHIRPSTACK_API_TOKEN="7F0as987e61..."  # ChirpStack API key
$ export JOIN_EUI="0101010102020203"            # Set JoinEUI for exported devices
$ export FREQUENCY_PLAN_ID="EU_863_870"         # Set FrequencyPlanID for exported devices
```

>Note: `JoinEUI` and `FrequencyPlanID` have to be set because ChirpStack does not store these variables.

## Export End Devices

With `ttn-lw-migrate` tool you can export a single or multiple end devices based on their `DevEUI`.

To export a single end device's description to a `device.json` file, use the following command in your terminal:

```bash
$ ttn-lw-migrate --source chirpstack device "0102030405060701" > device.json
```

To export multiple end devices, you need to create a `.txt` file containing one DevEUI per line as in example below.

<details><summary>Example of devices.txt</summary>

```bash
0102030405060701
0102030405060702
0102030405060703
0102030405060704
0102030405060705
0102030405060706
```

</details>

To export multiple end devices to a `devices.json` file, run the following command in your terminal:

```bash
$ ttn-lw-migrate --source chirpstack device < devices.txt > devices.json
```

## Export Applications

You can also export applications with `ttn-lw-migrate` tool using their names, which results with a JSON file containing descriptions of all the end devices that the application contains.

Use the following command to export end devices from a single application:

```bash
$ ttn-lw-migrate --source chirpstack application "app1" > application.json
```

To export end devices from multiple applications to an `applications.json` file, you need to create a `.txt` file containing one application name per line and run the following command in your terminal:

```bash
$ ttn-lw-migrate --source chirpstack application < applications.txt > applications.json
```

>**Notes**: 
>- ABP end devices without an active session can be exported from ChirpStack, but cannot be imported in {{% tts %}}.
>- `MaxEIRP` parameter may not be always set properly.
>- ChirpStack `variables` parameter related to payload formatting will always be converted to `null` when the end device is imported to {{% tts %}}.