---
title: "Setup"
description: ""
weight: 1
---

This section describes how to setup a Node-RED server and prepare to connect it to {{% tts %}}. 

## Requirements

1. [Install Node-RED](https://nodered.org/docs/getting-started/local)

>Note: Node-RED v1.0.6 is current at time of writing and is used in this guide.

## Setup

Run Node-RED and navigate to `http://localhost:1880` (or the public  address of your Node-RED instance). In your web browser, you should see something like:

{{< figure src="nodered_dashboard.png" alt="Node-RED dashboard" >}}

On the left side, you can see various types of nodes that can be used in order to build flows. All nodes can be found in the [Node-RED library](https://flows.nodered.org/).

{{% tts %}} Console provides the connection information needed for completing this integration. 

In the Console, click **Applications** and choose the application you want to connect to Node-RED. Click **Integrations** in the left hand panel of the Console, and the **MQTT** submenu to view the MQTT Server info:

{{< figure src="console_info.png" alt="MQTT Server connection information" >}}

In this example, the built-in MQTT Server is configured by default on port 1883 for insecure connections and on port 8883 for TLS-secured connections.

In a later step, we will use this information to connect Node-RED to {{% tts %}}.
