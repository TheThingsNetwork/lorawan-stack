---
title: "Creating a Webhook"
description: ""
weight: 2
---

Next, create a Webhook integration on {{% tts %}} by following this section.

<!--more-->

>Note: **TagoIO** Webhook template is now available on {{% tts %}}, but in case you want to create a **Custom webhook**, this guide can be helpful. Read more about these templates in the [Webhook templates]({{< ref "/integrations/webhooks/webhook-templates" >}}) page.

>Note: this section follows the [HTTP Webhooks]({{< ref "/integrations/webhooks" >}}) guide.  

Fill in the **Webhook ID** field. 

Create an `Authorization` header entry with the authorization copied in the previous step as a value.

Choose **JSON** as a **Webhook format**.

Set the **Base URL** value to `https://ttn.middleware.tago.io`.

Tick the box besides the uplink message type to enable this webhook for it and enter `/uplink` as an additional path to be appended to the **Base URL**.

{{< figure src="creating-a-webhook.png" alt="TagoIO webhook" >}}

After following these steps, you will see messages arriving in the **Live Inspector** tab in TagoIO.