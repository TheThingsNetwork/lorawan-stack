---
title: "Creating a Webhook"
description: ""
weight: 2
---

After finishing Datacake setup, make a Webhook integration on {{% tts %}} with these steps.

<!--more-->

>Note: **Datacake** Webhook template is now available on {{% tts %}}, but in case you want to create a **Custom webhook**, this guide can be helpful. Read more about these templates in the [Webhook templates]({{< ref "/integrations/webhooks/webhook-templates" >}}) page.

>Note: this section follows the [HTTP Webhooks]({{< ref "/integrations/webhooks" >}}) guide. 

Fill in the **Webhook ID** field and choose **JSON** for **Webhook format**. 

Next, you need to add an **Authorization** header, whose value will consist of the word "Token" and your [API token](https://docs.datacake.de/api/generate-access-token) from Datacake.

Per the **Webhook Settings** information that can be found in the **Configuration** tab on Datacake, set the **Base URL** value to `https://api.datacake.co/integrations/lorawan/tti/`.

{{< figure src="tts-datacake-webhook.png" alt="Datacake webhook" >}}

Check the message types for which you want to enable this webhook.

>Note: Datacake webhook template has the `Uplink message` type enabled by the default. 

Once the setup is finished, you can navigate to device's **Debug** tab on Datacake, where you can see the incoming messages and proceed with manipulating or monitoring your data.

Check the official Datacake documentation to learn how to [decode the payload](https://docs.datacake.de/lorawan/payload-decoders) received from {{% tts %}}. 