---
title: Export End Devices From V2
weight: 40
---

In this step, the end devices from {{% ttnv2 %}} will be exported in a JSON format that can then be parsed and imported by {{% tts %}}.

The exported end devices contain the device name, description, location data, activation mode (ABP/OTAA), root keys and the AppEUI. They also contain the session keys, so your OTAA devices can simply keep working with {{% tts %}}.

To get started, select the **AppID** and **AppEUI** of the application you want to export your end devices from:

```bash
$ ttnctl applications select
```

After selecting the application, make sure that you can list the available end devices:

```bash
$ ttnctl devices list
```

### Exporting Devices

In order to export a single device, use the following command. The device will be saved to `device.json`.

```bash
$ ttnctl devices export "device-id" --frequency-plan-id EU_863_870 > device.json
```

Alternatively, you can export all the end devices with a single command and save them in `all-devices.json`.

```bash
$ ttnctl devices export-all --frequency-plan-id EU_863_870 > all-devices.json
```

> **NOTE**: In the command above we used the `EU_863_870` frequency plan. You will need to change this to the frequency plan corresponding to your region. See [Frequency Plans]({{< ref "/reference/frequency-plans" >}}) for a list of supported Frequency Plans and their respective IDs.

> **NOTE**: Keep in mind that an end device can only be registered in one Network Server at a time. After importing an end device to {{% tts %}}, you should remove it from {{% ttnv2 %}}. For OTAA devices, it is enough to simply change the AppKey, so the device can no longer join but the existing session is preserved. Next time the device joins, the activation will be handled by {{% tts %}}.

### Disable Exported End Devices on V2

After exporting, make sure to clear the AppKey of your OTAA devices. This can be achieved with the following command:

```bash
$ ttnctl devices convert-to-abp "device-id" --save-to-attribute "original-app-key"
```

There is also a convenience command to clear all your devices at once:

```bash
$ ttnctl devices convert-all-to-abp --save-to-attribute "original-app-key"
```

> **NOTE**: The AppKey of each device will be printed on the standard output, and stored as a device attribute (with name `original-app-key`). You can retrieve the device attributes with `ttnctl devices info "device-id"`.
