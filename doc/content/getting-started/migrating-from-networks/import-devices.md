---
title: Import End Devices in The Things Stack
weight: 20
---

{{< cli-only >}}

To import your devices, you need an Application in {{% tts %}}. This can be done by following instructions for [Creating an Application]({{< ref "integrations/adding-applications" >}}).

To import devices, use the `devices.json` file you created in the previous step with `ttn-lw-cli`:

```bash
$ ttn-lw-cli end-devices create --application-id "application-id" < devices.json
```

This will import your devices on {{% tts %}}. In case any device fails, you will see a relevant error message at the end of the output.

If the import is successful, your devices are added to the list of end devices in your application.

{{< figure src="../successful-import.png" alt="successful-import" >}}

If the import was successful, your devices are added to the list of end-devices in your application.