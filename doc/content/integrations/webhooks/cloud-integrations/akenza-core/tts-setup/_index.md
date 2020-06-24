---
title: "Creating a Webhook"
description: ""
weight: 2
---

Once you have prepared the setup in Akenza Core, follow this section to create the Webhook integration.

<!--more-->

>Note: this section follows the [HTTP Webhooks]({{< ref "/integrations/webhooks" >}}) guide. 

Fill in the **Webhook ID** field and choose **JSON** for **Webhook format**. 

Paste the copied **HTTP Uplink URL** from Akenza Core in **Base URL** field.

In order to send uplink messages to Akenza Core, check the **Enabled** box next to the **Uplink message**. 

{{< figure src="creating-webhook.png" alt="Akenza Core webhook" >}}

After creating the integration, you will be able to see uplink messages in JSON format coming to Akenza Core if you navigate to **Inventory** &#8594; device &#8594; **DATA**.