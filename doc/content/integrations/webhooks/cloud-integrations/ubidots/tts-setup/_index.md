---
title: "The Things Stack Setup"
description: ""
weight: 2
---

This section contains the instructions for defining an uplink payload formatter and creating a Webhook integration on {{% tts %}}.

<!--more-->

## Defining an Uplink Payload Formatter

To make the previously deployed Ubidots function be able to interpret the payload coming from {{% tts %}}, payload needs to be decoded and stored into `decoded_payload` object of the `uplink_message` object. This is achieved by using an uplink payload formatter.

You can define an upload payload formatter per application or per device. 

In this guide, Javascript payload formatter type with the following parameter is being used:

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

but you need to adjust it according to the payload that your device is sending.

## Creating a Webhook Integration

>Note: this section follows the [HTTP Webhooks]({{< ref "/integrations/webhooks" >}}) guide. 

Fill in the **Webhook ID** field and choose **JSON** for **Webhook format**. 

Create a **Content-type** header with **application/json** value.

Paste the copied URL of your Ubidots function to **Base URL** field.

Check the uplink message type and select **Add Webhook** to finish creating an integration. 

{{< figure src="ubidots-webhook-creation.png" alt="Ubidots webhook" >}}

Once you have created the integration, navigate to **Devices** tab in Ubidots dashboard and select **Devices**. 

In the list of devices, you will be able to find your device, since it is being automatically added. Click on the device and you will see automatically added variables with their latest values.