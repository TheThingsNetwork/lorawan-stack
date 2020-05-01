---
title: Export end devices from V2
weight: 40
---

## Export end devices with ttnctl

We will export the end devices of this application in a JSON format that can then
be parsed and imported by {{% tts %}}.

For the commands below, we are assuming that an application on The Things Network
with AppID `v2-application` and AppEUI `0011223300112233`.

Select that AppID and AppEUI with:

```bash
$ ttnctl applications select
```

After selecting the application, make sure that you can list the available end devices:

```bash
$ ttnctl devices list
```

In order to export a single device, use the following command. The device will
be saved to `device.json`.

```bash
$ ttnctl devices export "DeviceID" --frequency-plan-id EU_863_870 > device.json
```

Alternatively, you can export all the end devices with a single command and save
them in `all-devices.json`.

```bash
$ ttnctl devices export "DeviceID" --frequency-plan-id EU_863_870 > all-devices.json
```

> **NOTE**: In {{% tts %}}, the MAC settings are configurable per end device. This
> means that all end devices need a frequency plan. Above, we specify
> the `EU_863_870` frequency plan, but this needs to be replaced with the
> frequency plan corresponding to your region. A list of supported frequency plan IDs in
> [the lorawan-frequency-plans Github repository](https://github.com/TheThingsNetwork/lorawan-frequency-plans/blob/master/frequency-plans.yml).

> **NOTE**: The exported end devices contain the device name, description, location
> data, activation mode (ABP/OTAA), root keys and the application AppEUI.

> **NOTE**: For OTAA end devices, the session keys are not preserved. After importing,
> the end device will need to send a new join request on the Network Server of {{% tts %}}.
