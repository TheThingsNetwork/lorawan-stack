---
title: "Send Messages"
description: ""
weight: 3
---

This section explains the process of setting up a flow which publishes messages to a certain topic that the MQTT Server is subscribed to.

Doing this schedules downlink messages to be sent to your end device. This section follows the example for publishing downlink traffic in [MQTT Server]({{< ref "/integrations/mqtt" >}}) guide.

## Configure MQTT Out Node

Place the **mqtt out** node on the dashboard. 

Configure the **Server** options with the same settings as in the [Receive Events and Messages]({{< ref "/integrations/node-red#receive-events-and-messages" >}}) section.

Set **Topic** to `v3/{application_id}/devices/{device_id}/down/push` to schedule downlink messages (as stated in [MQTT Server]({{< ref "/integrations/mqtt" >}}) guide). 

Choose a **QoS** from listed options and state whether you want the MQTT Server to retain messages. 

{{< figure src="mqtt_out_node_properties.png" alt="mqtt out node properties" >}}

## Configure Inject Node

Place the **inject** node on the dashboard. Double-click on the node to configure its properties. 

Choose **buffer** under **Payload** and enter the payload you wish to send. 

>Note: in this example, a downlink message with hexadecimal payload `00 2A FF 00` is to be sent, so here we define the **Payload** field as a corresponding array of byte values.  

Define the period between the automatic injections if you want them, or choose **none** for **Repeat** if you wish to inject messages manually.

{{< figure src="inject_node_properties.png" alt="inject node properties" >}}

## Configure Function Node and Deploy

Next, you have to configure a **function** node, which converts previously defined payload to a downlink message with Base64 encoded payload.

Place the **function** node with the following structure on dashboard:

```bash
return {
  "payload": {
    "downlinks": [{
      "f_port": 15,
      "frm_payload": msg.payload.toString("base64"),
      "priority": "NORMAL"
    }]
  }
}
```

Connect the nodes and click **Deploy**. If the setup is correct, below the **mqtt out** node **connected** status will be reported and downlink messages will begin sending to your end device.

{{< figure src="send_downlink_flow.png" alt="send downlink flow" >}}
