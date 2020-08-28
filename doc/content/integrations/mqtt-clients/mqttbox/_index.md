---
title: "MQTTBox"
description: ""
weight: 
---

[MQTTBox](http://workswithweb.com/mqttbox.html) is a cross platform application that makes it simple to create MQTT clients, virtual device networks, and load test MQTT devices and brokers. 

<!--more-->

In this guide, we show how to create an MQTT client for subscribing or publishing to topics used by {{% tts %}} [MQTT Server]({{< ref "/integrations/mqtt" >}}).

## Requirements

1. [MQTTBox app](http://workswithweb.com/html/mqttbox/downloads.html) installed on your system.

## Configuring MQTTBox

>Note: this section follows the examples in the [MQTT Server]({{< ref "/integrations/mqtt" >}}) guide.

In the MQTTBox application, click the **Create MQTT Client** button on the left to add a new MQTT client. Give a name to your MQTT client by filling in the **MQTT Client Name** field. Choose **mqtt / tcp** from the **Protocol** drop down list, and enter the **Public address** of your {{% tts %}} instance in the **Host** field. Do the same for the **Username** and **Password** fields.

{{< figure src="mqttbox-config.png" alt="MQTTBox client configuration" >}}

If you want to enable TLS for security, choose **mqtt / tls** as **Protocol** and use the **Public TLS address** of {{% tts %}} MQTT Server for **Host** instead. You will also have to upload at least the **CA file** associated with your {{% tts %}} deployment.

{{< figure src="mqttbox-tls-config.png" alt="MQTTBox TLS client configuration" >}}

You may leave the other fields empty. Click **Save** to save the configuration.

If the MQTTBox client successfully connects to {{% tts %}} MQTT Server, the **Connected** status button will be visible in the upper right. You can also use this button do disconnect. 

Once connected, you can proceed with configuring the publishers or subscribers. 

>Note: you can add multiple MQTT clients, as well as multiple publishers and subscribers within those clients.

## Subscribe to Upstream Traffic

To listen to messages being sent by your device and published by the {{% tts %}} MQTT Server, select **Add subscriber** and fill in the **Topic to subscribe** field with the name of topic you wish to subscribe to.

In this section, we use the `v3/{application id}/devices/{device id}/up` topic structure for listening to uplink messages. 

>Note: See [Subscribing to Upstream Traffic]({{< ref "/integrations/mqtt#subscribing-to-upstream-traffic" >}}) section of the MQTT Server guide for a list of available topics you can subscribe to.

Choose an appropriate **QoS** value from the list, click the **Subscribe** button and messages from {{% tts %}} will begin arriving shortly.

{{< figure src="subscribe.png" alt="Subscribe" >}}

## Schedule Downlink Messages

To send messages to your end device, you need to schedule a downlink message as described in the [Publishing Downlink Traffic]({{< ref "/integrations/mqtt#publishing-downlink-traffic" >}}) section of the MQTT Server guide. You can achieve this by adding and configuring a publisher in the MQTTBox application.

Fill in the **Topic to publish** field with `v3/{application id}/devices/{device id}/down/push` and replace the variables with their values from the {{% tts %}} Console.

>Note: see the [Publishing Downlink Traffic]({{< ref "/integrations/mqtt#publishing-downlink-traffic" >}}) section to learn about using `/replace` instead of `/push`.

Choose an appropriate **QoS** value from the drop down menu.

You can define the payload in multiple formats. In this guide, we use the **Strings / JSON / XML / Characters** for **Payload Type**.

To send an example downlink, paste the following content in the **Payload** field:

```json
{
  "downlinks": [{
    "f_port": 15,
    "frm_payload": "vu8=",
    "priority": "NORMAL"
  }]
}
```

Click the **Publish** button and a message with the hexadecimal payload `BE EF` will be scheduled for sending to your end device.

{{< figure src="publish.png" alt="Publish" >}}
