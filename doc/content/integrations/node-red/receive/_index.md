---
title: "Receive Events and Messages"
description: ""
weight: 2
---

This section follows the process of setting up a flow which subscribes and listens to the events and messages that are being published by the MQTT Server. 

## Configure MQTT In Node

Find the **mqtt in** node in the library section and place it on the flow dashboard. This node connects to the MQTT broker and subscribes to messages from the specified topic. Double-click on the node to configure its properties.

{{< figure src="mqtt_in_node_properties.png" alt="mqtt in node properties" >}}

In the **Server** dropdown menu, select **Add new mqtt-broker** and click on the button besides to edit it. 

{{< figure src="mqtt_in_node_connection.png" alt="Configuring MQTT Server connection information" >}}

In the **Connection** tab, under **Server**, provide the MQTT Server address from {{% tts %}} Console. The **Port** value depends on whether the connection is to be set up as a secure one or not. In this example, TLS is used, so the **Port** value is set to 8883. In this case you should also check the **Enable secure (SSL/TLS) connection** box.

{{< figure src="mqtt_in_node_security.png" alt="Configuring MQTT Server credentials" >}}

In the **Security** tab, enter the **Username** and **Password** according to the values in {{% tts %}} Console.

Under **Topic**, a topic to listen to can be specified. For testing purposes, we will set the value to "#" (all topics). The **QoS** value can be selected from the listed options, as well as the **Output**. In this example, we want to see each message as **a parsed JSON object**. A full list of topics that you can subscribe to is mentioned in [MQTT Server]({{< ref "/integrations/mqtt" >}}) guide.

Next, add two **debug** nodes to the flow and connect both of them to the **mqtt in** node. One debug node will listen to the events, while the other will listen to published messages. To make this happen, set the **Output** parameter for the first debug node configuration to **complete msg object** value, and to **msg.payload** for the other.

To listen to uplink messages coming from end devices, subscribe to `v3/{application_id}/devices/{device_id}/up`, where `{application_id}` and `{device_id}` are the application and device you wish to connect to.

{{< figure src="receive_uplink_flow.png" alt="Final flow scheme" >}}

After setup, click on **Deploy** in the upper right corner.

If the setup is correct, **connected** will be reported below the **mqtt in** node. In the upper right corner, click on the **debug** icon to see the event messages and their payloads in JSON format.
