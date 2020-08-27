---
title: Import End Devices in The Things Stack
weight: 20
---

To import your devices, you need an application in {{% tts %}}. This can be done by following instructions for [Creating an Application]({{< ref "integrations/adding-applications" >}}).

To import devices, use the `devices.json` file you created in the previous step.

{{< tabs/container "Console" "CLI" >}}

{{< tabs/tab "Console" >}}

Open the application you created and click the button **Import end devices**

{{< figure src="../import-end-devices.png" alt="import devices" >}}

Select `The Things Stack JSON` as the **Format** and upload the `devices.json` file.

{{< figure src="../upload-file.png" alt="upload devices.json file" >}}

Wait for the end devices to be successfully imported. In case any device fails, you see a relevant error message in the console.

{{< figure src="../operation-finished.png" alt="import finished" >}}

If the import was successful, your devices are added to the list of end-devices in your application.

{{< /tabs/tab >}}

{{< tabs/tab "CLI" >}}

You need the latest version of `ttn-lw-cli`, the CLI for {{% tts %}}.

{{< cli-only >}}

To import the devices.json file, run the following command in `ttn-lw-cli`.

```bash
$ ttn-lw-cli end-devices create --application-id "imported-application" < devices.json
```

This will import your devices on {{% tts %}}. In case any device fails, you see a relevant error message at the end of the output.

If the import was successful, your devices is added to the list of end-devices in your application.


{{< /tabs/tab >}}


{{< figure src="../successful-import.png" alt="successful-import" >}}

You can now start using your devices with {{% tts %}}!

> **Note:** Don't forget to manually copy-paste the Payload Formatters (if applicable) from {{% ttnv2 %}} to your {{% tts %}} application. See [Payload Formatters]({{% ref "integrations/payload-formatters" %}}) for more details.