---
title: Import End Devices in The Things Stack
weight: 50
---

## Import Devices using ttn-lw-cli

To import your devices, use the `devices.json` file you created
in the previous step with `ttn-lw-cli`:

```bash
$ ttn-lw-cli dev create --application-id "v3-application" < devices.json
```

This will import your devices on {{% tts %}}. In case
any device fails, you see a relevant error message at the end of the output.

> **ΝΟΤΕ**: After importing an end device to {{% tts %}}, you should remove it
> from {{% ttnv2 %}}. For OTAA devices, it is enough to simply change the
> AppKey, so the device can no longer connect but the existing session is preserved.
> Next time the device joins, it will connect to {{% tts %}}.
>
> Keep in mind that an end device can only be registered in one Network Server
> at a time.

You can now start using your devices and gateways with {{% tts %}}!
