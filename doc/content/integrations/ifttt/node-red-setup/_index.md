---
title: "Node-RED Setup"
description: ""
weight: 2
---

This section shows how to create a flow that will act like a mediator between {{% tts %}} and IFTTT. This flow will receive JSON messages from {{% tts %}} via the Webhooks integration, extract the decoded payload and send it as a payload of a separate HTTP POST request to IFTTT.

<!--more-->

Run Node-RED and open a new flow by clicking the **+** button in the upper right.

Place the **http in** node on the dashboard and double-click on it to configure it.

Select **POST** as a **Method**.

In the **URL** field enter the arbitrary path name, e.g. `join` if you are enabling the integration for the `Join accept` messages. Keep in mind that this path name also needs to be used when creating a Webhook on {{% tts %}}.

Click **Done** to finish.

{{< figure src="http-in-node.png" alt="Configuring HTTP input node" >}}

Next, add a **function** node to the dashboard. This node is used to define the structure of the HTTP POST request to be sent to IFTTT.

In the **Function** field of its configuration, paste the following code, adjust it according to your setup and select **Done**:

```bash
msg.url = "..." # Paste the URL copied from IFTTT in the previous step
msg.method = "POST";
msg.payload = {
    # Adjust according to your payload or leave empty
    'value1' : msg.payload.uplink_message.decoded_payload["temperature"],
    'value2' : msg.payload.uplink_message.decoded_payload["humidity"]
}
return msg;
```

{{< figure src="function-node.png" alt="Configuring function node" >}}

Place the **http request** node on the dashboard.

In its configuration, choose **set by msg.method** for a **Method** and select **Done**.

{{< figure src="http-request-node.png" alt="Configuring HTTP request node" >}}

To avoid timeouts of the HTTP requests originating from {{% tts %}}, the **http in** node needs to be connected to an **http response** node.

Add the **http response** node to the dashboard and configure it to reply to these requests with `200 OK`.

{{< figure src="http-response-node.png" alt="Configuring HTTP response node" >}}

Finally, add the **debug** node to the dashboard, configure it to display a **complete msg object** as an **Output** and finish by selecting **Done**.

Connect these nodes as shown on the picture below and click the **Deploy** button in the upper right corner. Use the debug window below this button to monitor the results.

{{< figure src="final-setup.png" alt="Node-RED setup" >}}
