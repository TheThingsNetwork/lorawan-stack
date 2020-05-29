---
title: "Connecting to Node-RED"
description: ""
---

[Node-RED](https://nodered.org/) is free, JavaScript-based development tool for visual programming, developed to ease the process of wiring together hardware devices, APIs and online services.

The MQTT server that is exposed by the Application Server can be connected to Node-RED. This kind of integration allows setting up a Node-RED flow that listens to events and messages sent by the end devices. Also, it is possible to send messages to the end devices.

<!--more-->

This guide contains the instructions for setting up these flows, but first make sure to properly [install the Node-RED](https://nodered.org/docs/getting-started/local).

>Note: in this guide, Node-RED v1.0.6 is used.

## Setup

First, you need to run the Node-RED by using the following command in terminal:

```bash
$ node-red 
```

If Node-RED is successfully launched, when navigating to `https://localhost:1880` in your web browser you should see something like:

{{< figure src="nodered_dashboard.png" alt="Node-RED dashboard" >}}

On the left side, you can see various types of nodes (they can also be found in the [Node-RED library](https://flows.nodered.org/)) that can be used in order to build flows, although in this guide only a few simple built-in nodes are used. 

{{% tts %}} Console provides the information needed for completing this integration, under **MQTT** submenu from **Integrations** menu within desired application.

{{< figure src="console_info.png" alt="MQTT Server connection information" >}}

In this example, since {{% tts %}} is running locally, the built-in MQTT Server is configured by default on port 1883 for insecure connections and on port 8883 for the TLS-secured ones.

## Receive Events and Messages

This section follows the process of setting up a flow which subscribes and listens to the events and messages that are being published by the MQTT Server. 

First, find the **mqtt in** node in the library section and place it on the flow dashboard. This node connects to the MQTT broker and subscribes to messages from the specified topic. Double-click on the node to configure its properties.

{{< figure src="mqtt_in_node_properties.png" alt="mqtt in node properties" >}}

In the **Server** dropdown menu, select **Add new mqtt-broker** and click on the button besides to edit it. 

{{< figure src="mqtt_in_node_connection.png" alt="COnfiguring MQTT Server connection information" >}}

In the **Connection** tab, under **Server**, you should provide MQTT Server's address. The **Port** value depends on whether the connection is to be set up as a secure one or not. In this example, TLS is being used, so the **Port** value is set to 8883. In this case you should also check the **Enable secure (SSL/TLS) connection** box.

{{< figure src="mqtt_in_node_security.png" alt="Configuring MQTT Server credentials" >}}

In the **Security** tab, enter the **Username** and **Password** according to values in {{% tts %}} Console.

Under **Topic**, a topic to listen to can be specified, but for the testing purposes set the value to "#" (all topics). The **QoS** value can be selected from the listed options, as well as the **Output**. In this example, we want to see each message as **a parsed JSON object**. Full list of topics that you can subscribe to is mentioned in [MQTT Server]({{< ref "/integrations/mqtt" >}}) guide.

Next, add two **debug** nodes to the flow and connect both of them to the **mqtt in** node. One debug node will listen to the events, while the other will listen to published messages. To make this happen, set the **Output** parameter for the first debug node configuration to **complete msg object** value, and to **msg.payload** for the other. In order to listen to messages coming just from the end devices, subscribe to `v3/{application id}/devices/{device id}/up`.

{{< figure src="receive_uplink_flow.png" alt="Final flow scheme" >}}

After setup, click on **Deploy** in the upper right corner. If the setup is correct, below the **mqtt in** node **connected** status will be reported. In upper right corner, click on debug icon to see the event messages and their payloads in JSON format.

## Send Messages 

This section follows the process of setting up a flow which publishes messages to a certain topic that MQTT Server is subscribed to. By doing this, you are scheduling downlink messages to be sent to your end device. This section follows the example for publishing downlink traffic in [MQTT Server]({{< ref "/integrations/mqtt" >}}) guide.

Find the **mqtt out** node in Node-RED library and place it on the dashboard. This node connects to MQTT broker and publishes messages. Configure the **Server** option same as in the upper section. As stated in [MQTT Server]({{< ref "/integrations/mqtt" >}}) guide, topic `v3/{application id}/devices/{device id}/down/push` can be used for scheduling downlink messages. Choose a **QoS** from listed options and state whether you want the MQTT Server to retain messages. 

{{< figure src="mqtt_out_node_properties.png" alt="mqtt out node properties" >}}

Further, find and place the **inject** node on the dashboard. This node injects a message into a flow and can be triggered manually, but it is also possible to define regular intervals between the automatic injections. Double-click on the **inject** node to configure its properties. **Payload** field can have different types of format - in this example JSON format is being used. Expand the **Payload** field and enter this:

```json
{
  "downlinks": [{
    "f_port": 15,
    "frm_payload": "vu8=",
    "priority": "NORMAL"
  }]
}
```

in order to schedule a downlink message with hexadecimal payload `BE EF`. State the period between the automatic injections if you want them or choose **None** for **Repeat** if you do not. 

Next, connect the nodes and click **Deploy**. If the setup is correct, below the **mqtt out** node **connected** status will be reported and you will start seeing the downlink messages being sent to your end device in {{% tts %}} Console. 

{{< figure src="send_downlink_flow.png" alt="mqtt out node properties" >}}