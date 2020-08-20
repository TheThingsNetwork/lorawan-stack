---
title: "Creating a Webhook"
description: ""
weight: 2
---

After finishing the setup in Azure, this section shows you how to create the Webhook integration on {{% tts %}}.

<!--more-->

>Note: this section follows the [HTTP Webhooks]({{< ref "/integrations/webhooks" >}}) guide. 

Fill in the **Webhook ID** field and choose **JSON** for **Webhook format**. 

Fill in the **Base URL** field with the function URL you copied from Azure in the previous step. 

Check the message types for which you wish to enable this webhook and finish creating an integration with **Add Webhook** button in the bottom. 

{{< figure src="azure-webhook-creation.png" alt="Azure webhook" >}}

After creating the integration, navigate to the logs console in Azure to see the incoming messages printed in JSON format.
