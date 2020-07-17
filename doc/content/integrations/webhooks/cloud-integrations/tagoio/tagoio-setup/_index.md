---
title: "TagoIO Setup"
description: ""
weight: 1
---

This section helps you to prepare TagoIO setup for integration with {{% tts %}}.

<!--more-->

Log in to your TagoIO user account and click the **Devices** button on the left hand menu. 

Select **Add Device** to add a new device.

{{< figure src="add-device.png" alt="Adding a device on TagoIO" >}}

The list of possible devices will pop out. Choose **HTTP** and then select **Custom HTTPS**.

{{< figure src="http-device.png" alt="Defining an HTTP device" >}}

Fill in the **Device name** field and click the **Create device** button to finish.

Select the created device in **Devices** to access its features.

On the **General Information** tab, find the **Token & Serial Number** section, where you can generate and copy a token needed in the next section of this guide.

{{< figure src="device-features.png" alt="Device features and token" >}}

Navigate to the **Live Inspector tab** and press the **Start** button to prepare the integration for the incoming messages from {{% tts %}}.