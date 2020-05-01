---
title: Import end devices in V3
weight: 50
---

## Import devices using ttn-lw-cli

In order to import your devices, just use the `devices.json` file you created
in the previous step with `ttn-lw-cli`:

```bash
$ ttn-lw-cli dev create --application-id "v3-application" < devices.json
```

If all goes well, this will import your devices on the {{% tts %}}. In case
any device fails, you see a relevant error message at the end of the output.

> **ΝΟΤΕ**: After importing an end device to {{% tts %}}, you should remove it
> from The Things Network. For OTAA devices, it is enough to simply change the
> AppKey, so the device can no longer but the existing session is preserved.
> Next time the device joins, it will be on the v3.
>
> Keep in mind that an end device can only be registered in one Network Server
> at a time.

You can now start using your devices and gateways with {{% tts %}}!
