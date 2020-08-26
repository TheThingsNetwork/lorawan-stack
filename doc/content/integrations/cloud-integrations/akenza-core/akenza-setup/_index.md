---
title: "Akenza Core Setup"
description: ""
weight: 1
---

This section shows how to prepare Akenza Core setup before creating a Webhook integration on {{% tts %}}.

<!--more-->

Log in to Akenza Core and navigate to the **Quick Start** tab in the **Environment** to see the steps to be taken to connect a device and manipulate the data coming from it. 

{{< figure src="quick-start.png" alt="Steps to connect a device" >}}

Set up a domain by selecting the **Add Domain** button in the **Domains** submenu. 

Give a **Name** to your domain and choose **HTTP** for **Technology** to create a Webhook integration. The **Domain Secret** is auto-generated, while **Uplink Function** and **Downlink Function** may retain a **Passthrough** value. 

Select **Save** to finish.

{{< figure src="creating-domain.png" alt="Creating a new domain" >}}

Next, create a new device type in the **Device Types** submenu. Select the **Add Device Type** button, provide a **Name** and click **Save**.

{{< figure src="creating-device-type.png" alt="Creating a device type" >}}

After defining a domain and a device type, you can create a new device by selecting **Add Device** button in the **Inventory** submenu. 

Provide a **Name** for your device, select the previously created device type and domain, and do not forget to generate a **Device ID**.

{{< figure src="creating-device.png" alt="Creating a new device" >}}

Once the device is created, select it in **Inventory** and scroll down to the bottom of the page. Under **Endpoints**, you will find the **HTTP Uplink URL** that you need for the Webhook integration.
