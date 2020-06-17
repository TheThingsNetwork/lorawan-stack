---
title: "Creating a Webhook"
description: ""
weight: 2
---

After finishing Datacake setup, make a Webhook integration on {{% tts %}} with these steps.

<!--more-->

>Note: this section follows the [HTTP Webhooks]({{< ref "/integrations/webhooks" >}}) guide. 

Fill in the **Webhook ID** field and choose **JSON** for **Webhook format**. 

Next, you need to add an **Authorization** header, whose value will consist of the word "Token" and your [API token](https://docs.datacake.de/api/generate-access-token) from Datacake.

Per the **Webhook Settings** information that can be found in the **Configuration** tab on Datacake, set the **Base URL** value to `https://api.datacake.co/integrations/lorawan/tti/`.

{{< figure src="tts-datacake-webhook.png" alt="Datacake webhook" >}}

Check the message types for which you want to enable this webhook.

Once the setup is finished, you can navigate to device's **Debug** tab, where you can see the incoming messages and their details.
