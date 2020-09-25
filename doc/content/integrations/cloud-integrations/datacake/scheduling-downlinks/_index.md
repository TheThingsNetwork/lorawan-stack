---
title: "Scheduling Downlinks with Datacake"
description: ""
weight: 3
---

Besides forwarding messages from {{% tts %}} to Datacake, you can also schedule downlink messages to be sent from Datacake towards your end device.

<!--more-->

Enter your device's settings page on Datacake and go to **Downlinks** tab.

Click the **Add Downlink** button.

Next, fill in the **Name** field, define the **Payload encoder** and **Save Downlink**. 

>Note: learn how to write downlink payload decoder with [this guide](https://docs.datacake.de/lorawan/downlinks#writing-a-downlink-encoder) from the official Datacake documentation site.

{{< figure src="downlink-configuration.png" alt="Configuring downlink" >}}

Now simply click the **Send Downlink** button to schedule a downlink and check your device's logs to see the incoming message.