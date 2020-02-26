---
title: "Create Gateway"
description: ""
weight: 2
---

Learn how to register a gateway using the Console.

<!--more-->

## Step by step

Go to **Gateways** in the top menu, and click **+ Add Gateway** to reach the gateway registration page. Fill the gateway ID, gateway EUI (if your gateway has an EUI) and frequency plan. The other fields are optional. Click **Create Gateway** to create the gateway.

{{< figure src="gateway-creation.png" alt="Gateway creation" >}}

Your gateway will be created and you will be redirected to the gateway overview page of your newly created gateway.

{{< figure src="gateway-overview.png" alt="Gateway overview" >}}

You can now connect your gateway to {{% tts %}}.

### Create Gateway API Key

Some gateways require an API Key with Link Gateway Rights to be able to connect to {{% tts %}}. 

In order to do this, navigate the **API Keys** menu of your gateway and select **Add API Key**. Enter a name for your key, select the **Link as Gateway to a Gateway Server for traffic exchange, i.e. write uplink and read downlink** right and then press **Create API Key**.

{{< figure src="gateway-api-key-creation.png" alt="Gateway API Key creation" >}}

You will see a screen that shows your newly created API Key. You now can copy it in your clipboard by pressing the clipboard button. After saving the key in a safe place, press **I have copied the key**. You will not be able to see this key again in the future, and if you lose it, you can create a new one to replace it in the gateway configuration.

{{< figure src="gateway-api-key-created.png" alt="Gateway API Key created" >}}
