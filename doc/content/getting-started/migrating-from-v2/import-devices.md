---
title: Import End Devices in The Things Stack
weight: 50
---

{{< cli-only >}}

To import your devices to the application you [created]({{% relref "create-application" %}}), use the `devices.json` file you created in the previous step with `ttn-lw-cli`:

```bash
$ ttn-lw-cli end-devices create --application-id "imported-application" < devices.json
```

This will import your devices on {{% tts %}}. In case any device fails, you see a relevant error message at the end of the output.

If the import was successful, your devices is added to the list of end-devices in your application.

{{< figure src="../successful-import.png" alt="successful-import" >}}

You can now start using your devices with {{% tts %}}!
