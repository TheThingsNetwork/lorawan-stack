---
title: "Creating a Webhook"
description: ""
weight: 3
---

Follow this section to create a Webhook integration with the **http in** node from Node-RED.

<!--more-->

>Note: this section follows the [HTTP Webhooks]({{< ref "/integrations/webhooks" >}}) guide. 

Give a name to your webhook by filling in the **Webhook ID** field. 

For the **Webhook format**, choose **JSON**.

Enter the **Base URL** value according to your Node-RED deployment.

Select the message type you want to enable this webhook for and fill in the path to be appended to the **Base URL** accordingly. Keep in mind that this path needs to be the same as the path provided in **http in** node. 

Finish by clicking the **Add webhook** button.

{{< figure src="creating-a-webhook.png" alt="Creating a webhook" >}}

Once you have completed the integration, your IFTTT applet can be triggered by events on {{% tts %}} and the payload values from {{% tts %}} messages can be incorporated in the actions defined in your applet. The setup shown in this guide leads to receiving an SMS containing temperature and humidity sensor values whenever a join request from the device is accepted.