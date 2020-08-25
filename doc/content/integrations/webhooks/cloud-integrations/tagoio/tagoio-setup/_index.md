---
title: "TagoIO Setup"
description: ""
weight: 1
---

This section helps you to prepare TagoIO setup for integration with {{% tts %}}.

<!--more-->

Log in to your TagoIO user account and click the **Devices** button on the left hand menu. 

Select **Add Device** in the upper right to add a new device.

The list of available devices will pop out. Choose **LoRaWAN TTN** and then select **Custom The Things Network**.

{{< figure src="custom-ttn-device.png" alt="Choosing a Custom TTN device" >}}

Give a name to your device by filling the **Device name** field, enter the **Device EUI** and click the **Create device** button to finish.

{{< figure src="device-settings.png" alt="Configuring a device" >}}

When your device has been created, the window with a note for generating an [authorization](https://docs.tago.io/en/articles/218) will pop out. Click the **Generate Authorization** button. 

{{< figure src="auth-pop-out.png" alt="Generate Authorization pop-out window" >}}

When redirected to the **Service Authorization** page, fill in the **Name** field and select **Generate**. Copy this value for further steps.

Select the created device in **Devices** to access its features.

{{< figure src="device-features.png" alt="Device features" >}}

Navigate to the **Live Inspector** tab and press the **Start** button to prepare the integration for the incoming messages from {{% tts %}}.