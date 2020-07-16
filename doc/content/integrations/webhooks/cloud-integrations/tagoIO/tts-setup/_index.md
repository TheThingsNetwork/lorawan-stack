---
title: "Creating a Webhook"
description: ""
weight: 2
---

Next, create a Webhook integration on {{% tts %}} by following this section.

<!--more-->

>Note: this section follows the [HTTP Webhooks]({{< ref "/integrations/webhooks" >}}) guide. 

Fill in the **Webhook ID** field. 

Create a `Content-Type` header entry with `application/json` value, and a `Device-Token` header entry with the token value you previously copied from TagoIO.

Choose **JSON** for **Webhook format**

Set the **Base URL** value to `https://api.tago.io/data`.

Tick the box besides the message types which you want to enable this webhook for and select **Add webhook**.

{{< figure src="creating-a-webhook.png" alt="TagoIO webhook" >}}

After following these steps, you will see messages arrive in the **Live Inspector** tab in TagoIO.