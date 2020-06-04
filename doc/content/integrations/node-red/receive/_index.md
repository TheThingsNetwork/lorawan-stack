---
title: "Receive Events and Messages"
description: ""
weight: 2
---

This section follows the process of setting up a flow which subscribes and listens to the events and messages that are being published by the MQTT Server. 

## Configure MQTT In Node

1. Place the **mqtt in** node on the dashboard. Double-click on the node to configure its properties.

2. In the **Server** dropdown menu, select **Add new mqtt-broker** and click on the button besides to edit it. 

3. In the **Connection** tab, under **Server**, provide the MQTT Server address from {{% tts %}} Console. 

>Note: in this example, TLS-secured connection is to be established, so the **Port** value is set to 8883. In this case you should also check the **Enable secure (SSL/TLS) connection** box.

{{< figure src="mqtt_in_node_connection.png" alt="Configuring MQTT Server connection information" >}}

4. In the **Security** tab, enter the **Username** and **Password** according to the values in {{% tts %}} Console.

{{< figure src="mqtt_in_node_security.png" alt="Configuring MQTT Server credentials" >}}

5. Go back to **Properties** and set the **Topic** value to `#` (to subscribe to all topics). 

>Note: a full list of topics that you can subscribe to is mentioned in [MQTT Server]({{< ref "/integrations/mqtt" >}}) guide. 

6. Select the **QoS** value from the listed options and set **Output** parameter to **a parsed JSON object**. 

## Configure Debug Nodes

1. Add two **debug** nodes and connect both to the **mqtt in** node. One debug node will listen to the events, while the other will listen to published messages. 

>Note: you can also subscribe to `v3/{application_id}/devices/{device_id}/up` to only listen to uplink messages coming from end devices, as mentioned in [MQTT Server]({{< ref "/integrations/mqtt" >}}) guide.

2. Set the **Output** parameters of these nodes to **complete msg object** and **msg.payload**.

## Deploy

1. Click on **Deploy** in the upper right corner. If the setup is correct, **connected** status will be reported below the **mqtt in** node. 

2. Click on **debug** icon in the upper right corner to see the published event messages and their payloads in JSON format.

{{< figure src="receive_uplink_flow.png" alt="Final flow scheme" >}}