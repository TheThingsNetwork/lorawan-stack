---
title: "MAC Settings"
description: ""
---

This section provides guidelines for configuring MAC settings for end devices from the CLI.

<!--more-->

{{< cli-only >}}

### MAC Settings and MAC State

MAC settings on {{% tts %}} are configurable per end device. To configure persistent MAC settings, make changes to `mac-settings.desired_<parameter>`. Updates to `mac-settings.desired_<parameter>` take effect on device creation, after OTAA join or ABP FCnt reset, ResetInd, or after MAC state reset.

`mac-settings.<parameter>` represents what the Network Server believes is configured on the end device, and should not be changed, unless the device does not conform to spec. It may however be necessary to set `mac-settings.RX1_delay` for ABP devices where this is not configured as part of activation.

`mac-state` can be used to test MAC settings in the current session. To update settings for testing in the current session, make changes to the `mac-state.desired_parameters.<parameter>`. Updates to the `mac-state.desired_parameters.<parameter>` are applied on the next uplink, and lost on reset.

The expected procedure for testing and updating settings is:

1. Modify `mac-state.desired_parameters.<parameter>` to see changes in the current session
2. Test that everything works as expected
3. Modify `mac-settings.desired_<parameter>` to make the change permanent

If no settings are provided on device creation or unset, defaults are first taken from the device Frequency Plan if available, and finally from [Network Server Configuration]({{< ref src="/reference/configuration/network-server" >}}).

### Available MAC settings

Run the following command to get a list of all available MAC settings and available parameter values:

```bash
$ ttn-lw-cli end-devices set --help
```

You can also refer to the [End Device API Reference page]({{< ref "reference/api/end_device#message:MACSettings" >}}) for documentation on the available MAC settings and MAC state parameters.

### Class Specific Settings

Settings that are useful based on device class are:

All devices:

- `mac-settings.factory-preset-frequencies`

Class A:

- `mac-settings.desired-rx1-delay`
- `mac-settings.desired-rx1-data-rate-offset`
- `mac-settings.desired-rx2-data-rate-index`
- `mac-settings.desired-rx2-frequency`
- `mac-settings.supports-32-bit-f-cnt`
- `mac-settings.use-adr`

Class A ABP:

- `mac-settings.resets-f-cnt`

Class B:

- `mac-settings.class-b-timeout`
- `mac-settings.ping-slot-periodicity`
- `mac-settings.desired-ping-slot-data-rate-index`
- `mac-settings.desired-ping-slot-frequency`

Class C:

- `mac-settings.class-c-timeout`

Some additional examples are included below. All settings are available at the [End Device API Reference page]({{< ref "reference/api/end_device#message:MACSettings" >}}) and can be viewed using the `ttn-lw-cli end-devices set --help` command.

### Configure Factory Preset Frequencies for ABP Devices

To tell {{% tts %}} which frequencies are configured in an ABP device, set the `mac-settings.factory-preset-frequencies` parameter. For example, to configure a device using the default EU868 frequencies, use the following command:

```bash
$ ttn-lw-cli devices update <device_id> --mac-settings.factory-preset-frequencies 868100000,868300000,868500000,867100000,867300000,867500000,867700000,867900000
```

>Note: For ABP devices, `mac-settings.factory-preset-frequencies` should be specified on `device create` or the settings will only take effect after MAC reset.

### Set Duty Cycle

To change the duty cycle, set the `desired-max-duty-cycle` parameter. For example, to set the duty cycle to 0.098%, use the following command:

```bash
$ ttn-lw-cli end-devices set <app-id> <device-id> --mac-settings.desired-max-duty-cycle DUTY_CYCLE_1024
```

>Note: See the [End Device API Reference]({{< ref "reference/api/end_device#message:MACSettings" >}}) for available fields and definitions of constants. DUTY_CYCLE_1024 represents 1/1024 ≈ 0.098%.

### Enable ADR

To enable ADR, set the `mac-settings.use-adr` parameter

```bash
$ ttn-lw-cli end-devices set <app-id> <device-id> --mac-settings.use-adr true 
```

### Set RX1 Delay

The RX1 delay of end devices is set to 5 second by default. To change it, set the `desired-rx1-delay` parameter:

```bash
$ ttn-lw-cli end-devices set <app-id> <device-id> --mac-settings.desired-rx1-delay RX_DELAY_5
```

### Unset MAC settings

The CLI can also be used to unset MAC settings (so that the default ones are used):

```bash
$ ttn-lw-cli end-devices set <app-id> <device-id> --unset mac-settings.rx1-delay
```
