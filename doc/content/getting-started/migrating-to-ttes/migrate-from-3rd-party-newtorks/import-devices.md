---
title: Import End Devices in The Things Stack
weight: 20
---

{{< cli-only >}}

To import your devices to the application you [created]({{% ref "getting-started/migrating-to-ttes/preparing-ttes-environment/create-application" %}}), use the `devices.json` file you created in the previous step with `ttn-lw-cli`:

```bash
$ ttn-lw-cli end-devices create --application-id "imported-application" < devices.json
```

This will import your devices on {{% tts %}}. In case any device fails, you see a relevant error message at the end of the output.

You can now start using your devices and gateways with {{% tts %}}!





