---
title: "Ursalink UG8X IoT LoRaWAN Gateway"
description: ""
---

The **Ursalink UG8X IoT LoRaWAN Gateway** is an 8 channel (16 channel optional) configurable, scalable gateway for industrial IoT applications.

This page contains information about connecting the Ursalink UG8X IoT LoRaWAN Gateway to {{% tts %}}

<!--more-->

The technical specifications can be found in [Ursalink's official documentation](https://www.ursalink.com/en/ad-lorawan-gateway/). The Ursalink UG8X IoT LoRaWAN Gateway supports two ways to connect with {{% tts %}}, using either the Semtech Packet Forwarder or {{% lbs %}}.

{{< figure src="ursalink.jpg" alt="Ursalink">}}

## Requirements

1. User account on {{% tts %}} with rights to create gateways.
2. Ursalink UG8X LoRaWAN Gateway connected to the internet via ethernet or cellular backhaul.
3. CA certificate for {{% lbs %}} (if using {{% lbs %}}).

## Registration

Create a gateway by following the instructions for the [Console]({{< ref "/getting-started/console#create-gateway" >}}) or the [CLI]({{< ref "/getting-started/cli#create-gateway" >}}).

The **EUI** of the gateway can be found on the configuration web page of the gateway. See the [next section]({{< ref "#configuration-via-browser" >}}) for instructions to access the configuration page.

{{< figure src="eui.png" alt="Gateway EUI" >}}

## Configuration via Browser

Find the IP address of the gateway. The default IP for the Ursalink UG8X LoRaWAN Gateway is 192.168.23.150.

Connect your machine to the same local network as that of the gateway, and enter the IP address in your web browser. The default username is **admin** and the default password is **password**. See [Ursalink's official documentation](https://www.ursalink.com/en/ad-lorawan-gateway/) for more information.

{{< figure src="login.png" alt="Login" >}}

### Disable Default Server

In the left menu, choose **Packet Forwarder**. Select the **General** tab.

{{< figure src="eui.png" alt="Packet Forwarder" >}}

Click the pencil icon next to the default server, and uncheck the **Enabled** button to disable the default server.

Click **Save** to continue.

{{< figure src="disable.png" alt="Disable default server" >}}

## Connect to {{% tts %}}

After completing basic configuration, follow the instructions for connecting using [LBS]({{< relref "lbs" >}}) or the [UDP Packet Forwarder]({{< relref "packet-forwarder" >}}).
