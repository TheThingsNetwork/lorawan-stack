---
title: Fine-tuning MAC Settings for End Devices
weight: 60
---

{{< cli-only >}}

This section is relevant for end devices that need MAC settings fine-tuning to function properly with the Network Server of {{% tts %}}.

MAC settings on {{% tts %}} are configurable per end device. They can be configured from the CLI.

### Setting RX1 Delay

The RX1 delay of end devices is set to 1 second by default. For some end devices, this may lead to downlink messages not being scheduled in time. Therefore, it is recommended that the RX1 delay be increased to 5 seconds:

```bash
$ ttn-lw-cli end-devices set "app-id" "device-id" --mac-settings.rx1-delay RX_DELAY_5
```

### Unsetting MAC settings

The CLI can also be used to unset MAC settings (so that the default ones are used):

```bash
$ ttn-lw-cli end-devices set "app-id" "device-id" --unset mac-settings.rx1-delay
```

### Available MAC settings

Run the following command to get a list of all available MAC settings:

```bash
$ ttn-lw-cli end-devices set --help
```

You can also refer to the [API Reference page]({{< ref "reference/api/end_device#message:MACSettings" >}}) for documentation on the available MAC settings and MAC state parameters.

<!--
TODO: https://github.com/TheThingsNetwork/lorawan-stack/issues/2501
Update with reference to the MAC settings page
-->
