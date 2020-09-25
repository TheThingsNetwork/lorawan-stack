---
title: "Datacake Setup"
description: ""
weight: 1
---

Follow the instructions in this section to prepare Datacake setup for integration with {{% tts %}}.

<!--more-->

First, create a **Workspace** on Datacake by navigating to the **Create Workspace** button in the upper left corner. 

{{< figure src="datacake-workspace.png" alt="Add workspace on Datacake" >}}

On the left hand menu, click **Devices**. To add a new device, click the **Add Device** button on the right. 

In the **Add Device** pop-up menu, choose **LoRaWAN** &#8594; **Generic LoRa Device**. When asked **Which device are you missing?**, just click **Skip** and then select **The Things Industries** adapter.

{{< figure src="datacake-tti-adapter.png" alt="Datacake TTI adapter" >}}

After selecting your subscription plan, fill in **Name** and **DevEUI** for your device, then click on **Add Device** to finish.

Once the device is created, you can click on it in the **Devices** menu to enter its settings.

In the **Configuration** tab, you can find **LORaWAN** section, where you can configure **Network** settings, choose to **Authenticate Webhook** or define **Payload Decoder**. 

Under **Network**, choose **The Things Industries** from the drop-down menu.

Paste your device's **End device ID** from {{% tts %}} in the **TTI Dev Id** field.

**TTI Server Url** field should contain the URL of your {{% tts %}} deployment.

Paste your **Application ID** from {{% tts %}} in the **TTI App Id** field. 

In {{% tts %}}, navigate to **API keys** on the left hand menu, click the **Add API key** button, give it a **Name** and confirm that you have copied it to finish. Paste the copied API key into the **TTI Api Key** field on Datacake. 

{{< figure src="lorawan-settings.png" alt="LoRaWAN settings" >}}

