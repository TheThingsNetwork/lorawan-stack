---
title: "Datacake Setup"
description: ""
weight: 1
---

Follow the instructions in this section to prepare Datacake setup for integration with {{% tts %}}.

<!--more-->

First, create a **Workspace** on Datacake by navigating to the **Create Workspace** button in the upper left corner. 

{{< figure src="datacake-workspace.png" alt="Add workspace on Datacake" >}}

On the left hand menu, click **Devices**. To add a new device, click on **Add Device** button on the right. 

In the **Add Device** pop-up menu, choose **LoRaWAN** &#8594; **Generic LoRa Device** &#8594; **The Things Industries** adapter.

{{< figure src="datacake-tti-adapter.png" alt="Datacake TTI adapter" >}}

Fill in **Name** and **DevEUI** for your device, then click on **Add Device** to finish.

Once the device is created, you can click on it in the **Devices** menu to enter its settings.

In the **Configuration** tab, you can find **Payload Decoder**, which is intended for [decoding the payload](https://docs.datacake.de/lorawan/payload-decoders) received from {{% tts %}} after deploying the integration. In order to store the payload values you need to [create a field](https://docs.datacake.de/device/database/fields) according to the payload type. 
