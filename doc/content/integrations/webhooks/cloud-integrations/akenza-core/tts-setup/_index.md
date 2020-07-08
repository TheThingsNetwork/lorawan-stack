---
title: "Creating a Webhook"
description: ""
weight: 2
---

Once you have prepared the setup in Akenza Core, follow this section to create the Webhook integration.

<!--more-->

>Note: this section follows the [HTTP Webhooks]({{< ref "/integrations/webhooks" >}}) guide. 

Fill in the **Webhook ID** field and choose **JSON** for **Webhook format**. 

Paste the copied **HTTP Uplink URL** from Akenza Core in the **Base URL** field.

To send uplink messages to Akenza Core, check the **Enabled** box next to the **Uplink message**. 

{{< figure src="creating-webhook.png" alt="Akenza Core webhook" >}}

After creating the integration, you will be able to see uplink messages in JSON format in Akenza Core if you navigate to the **Data** tab of the created device (which can be found in the **Inventory**).