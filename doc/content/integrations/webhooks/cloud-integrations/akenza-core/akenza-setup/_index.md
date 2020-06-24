---
title: "Akenza Core Setup"
description: ""
weight: 1
---

This section shows how to prepare Akenza Core setup before creating a Webhook integration on {{% tts %}}.

<!--more-->

By logging in to your Akenza Core user account and navigating to **QUICK START** tab in the **Environment**, you can see the order of steps to be taken to connect a device and manipulate the data coming from it. 

{{< figure src="quick-start.png" alt="Steps to connect a device" >}}

First step is setting up a domain, which can be done by selecting the **ADD DOMAIN** button in **Domains** submenu. 

Give a **Name** to your domain and choose **HTTP** for **Technology** in order to implement a Webhook integration. **Domain Secret** is auto-generated, while **Uplink Function** and **Downlink Function** may retain **Passthrough** value. 

Select **SAVE** to finish.

{{< figure src="creating-domain.png" alt="Creating a new domain" >}}

In a similar manner, you need to create a new device type in **Device Types** submenu, where you just need to select **ADD DEVICE TYPE** button, provide a **Name** and click **SAVE**.

{{< figure src="creating-device-type.png" alt="Creating a device type" >}}

When you have defined a domain and a device type, you can create a new device by selecting **ADD DEVICE** button in **Inventory** submenu. 

Provide a **Name** for your device, select previously created device type and domain, and do not forget to generate **Device ID**.

{{< figure src="creating-device.png" alt="Creating a new device" >}}

Once the device is created, select it in **Inventory** and scroll down to the bottom of the page. Under **Endpoints** you can find and copy **HTTP Uplink URL** that you need for the Webhook integration.