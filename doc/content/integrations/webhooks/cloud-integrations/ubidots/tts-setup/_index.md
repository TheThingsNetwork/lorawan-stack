---
title: "The Things Stack Setup"
description: ""
weight: 2
---

This section contains the instructions for defining an uplink payload formatter and creating a Webhook integration on {{% tts %}}.

<!--more-->

## Defining an Uplink Payload Formatter

To allow our Ubidots function to interpret the payload coming from {{% tts %}}, the payload needs to be decoded. To do this, we use an uplink payload formatter which decodes the payload and stores the results in a `decoded_payload` field of the `uplink_message` object. 

For more information on message payload formatters, see [Payload Formatters]({{< ref "/integrations/payload-formatters" >}}).

For this guide, we will use the following example JavaScript payload formatter, which converts a byte encoded temperature and humidity to human readable fields.

```
function Decoder(bytes, fport) {
  var decoded = {};

  var temp = (bytes[0] << 8) | bytes[1];
  var hum = (bytes[2] << 8) | bytes[3];
  
  decoded.temperature = temp / 100;
  decoded.humidity = hum / 100;

  return decoded;
}
```

You should adjust the decoded based on the payload that your device is sending.

## Creating a Webhook Integration

>Note: this section follows the [HTTP Webhooks]({{< ref "/integrations/webhooks" >}}) guide. 

Fill in the **Webhook ID** field and choose **JSON** for **Webhook format**. 

Create a **Content-type** header with **application/json** value.

Paste the copied URL of your Ubidots function to **Base URL** field.

Check the uplink message type and select **Add Webhook** to finish creating an integration. 

{{< figure src="ubidots-webhook-creation.png" alt="Ubidots webhook" >}}

Once you have created the integration, navigate to **Devices** tab in Ubidots dashboard and select **Devices**. 

You should see your device listed, as it is automatically added when an uplink is received. Click the device to see variables and their latest values.