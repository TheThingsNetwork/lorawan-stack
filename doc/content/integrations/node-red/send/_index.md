---
title: "Send Messages"
description: ""
weight: 3
---

This section follows the process of setting up a flow which publishes messages to a certain topic that the MQTT Server is subscribed to.

Doing this schedules downlink messages to be sent to your end device. This section follows the example for publishing downlink traffic in [MQTT Server]({{< ref "/integrations/mqtt" >}}) guide.

## Configure MQTT Out Node

Find the **mqtt out** node in the Node-RED library and place it on the dashboard. This node connects to the MQTT broker and publishes messages. Configure the **Server** options with the same settings as used in the [Receive Events and Messages]({{< ref "/integrations/node-red#receive-events-and-messages" >}}) section before.

As stated in [MQTT Server]({{< ref "/integrations/mqtt" >}}) guide, topic `v3/{application_id}/devices/{device_id}/down/push` can be used for scheduling downlink messages. Choose a **QoS** from listed options and state whether you want the MQTT Server to retain messages. 

{{< figure src="mqtt_out_node_properties.png" alt="mqtt out node properties" >}}

Further, find and place the **inject** node on the dashboard. This node injects a message into a flow and can be triggered manually, but it is also possible to define regular intervals between the automatic injections. Double-click on the **inject** node to configure its properties. The **Payload** field can have different formats - in this example, JSON is used. Expand the **Payload** field and enter this:

```json
{
  "downlinks": [{
    "f_port": 15,
    "frm_payload": "vu8=",
    "priority": "NORMAL"
  }]
}
```

in order to schedule a downlink message with base64 encoded payload `vu8=`. State the period between the automatic injections if you want them or choose **None** for **Repeat** if you do not. 

Next, connect the nodes and click **Deploy**. If the setup is correct, below the **mqtt out** node **connected** status will be reported and downlink messages will begin sending to your end device.

{{< figure src="send_downlink_flow.png" alt="mqtt out node properties" >}}
