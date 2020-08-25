---
title: "Adding Webhook Templates"
description: ""
weight: 4
---

{{% tts %}} uses webhook templates from the [`lorawan-webhook-templates` Github repository](https://github.com/TheThingsNetwork/lorawan-webhook-templates/).

Once you have created a new webhook template with a proper [format]({{< ref "/integrations/webhooks/webhook-templates/format.md" >}}), you can easily test it locally by following the next steps:

1. Clone the [`lorawan-webhook-templates` Github repository](https://github.com/TheThingsNetwork/lorawan-webhook-templates/) to a local folder.

2. Store your webhook template in the previously mentioned folder.

3. Include your webhook template in the `templates.yml` file.

4. Update your {{% tts %}} [configuration]({{< ref "/getting-started/installation/configuration" >}}) file by adding the following lines:

```yaml
as:
  webhooks:
    templates:
      directory: "path-to-the-folder-containing-your-webhook-template"
```

or use `--as.webhooks.templates.directory` command line option when running {{% tts %}} instead.

Go to the Console and select **Webhooks** tab in **Integrations** menu. Click the **Add webhook** button and you will see your template. At this point, you can test your webhook template.

{{< figure src="../adding-webhook-template.png" alt="Webhook template successfully added" >}}

To make your webhook template available in most deployments when the next version is deployed, open a pull request on the [`lorawan-webhook-templates` Github repository](https://github.com/TheThingsNetwork/lorawan-webhook-templates/).