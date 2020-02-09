---
title: "MikroTik Routerboard wAP LoRa8 kit"
description: ""
weight: 1
---

MikroTik Routerboard wAP LoRa8 kit is a LoRaWAN gateway that contains a pre-installed UDP packet forwarder and an outdoor weatherproof wireless access point with 2.4 GHz WLAN interface and Ethernet port that could be used as a backend. It's technical specifications can be found in [the official documentation](https://mikrotik.com/product/wap_lora8_kit). 

There are multiple interfaces to configure the gateway parameters. This page guides you to connect it to {{% tts %}} using the [MikroTik Mobile App](https://mikrotik.com/mobile_app) for Android/iOS.

## Physical Connections

The MikroTik Routerboard wAP LoRa8 kit comes with a PoE adapter. The following image shows the proper connections to the device.

{{< figure src="mikrotik-routerboard-connections.jpg" alt="MikroTik Routerboard wAP LoRa8 kit connections" >}}

## Prerequisites

1. User account on {{% tts %}} with rights to create Gateways.
2. MikroTik Routerboard wAP LoRa8 kit connected via ethernet.
3. [MikroTik Mobile App](https://mikrotik.com/mobile_app) installed on a smartphone.

## Registration

Create a gateway by following the instructions for the [Console]({{< ref "/guides/getting-started/console#create-gateway" >}}) or the [CLI]({{< ref "/guides/getting-started/cli#create-gateway" >}}). The **EUI** of the gateway can be found on the back panel of the gateway under the field **GW ID**.

## Configuration

The MikroTik Routerboard exposes a WiFi Access Point (AP) with SSID `MikroTik-xxxxxx`, where `xxxxxx` are the last 6 digits of the device's MAC Address (printed on the back panel).

Using the device where the **MikroTik Mobile App** is installed, connect to this AP. By default, there's no password.

Open the **MikroTik Mobile App**. The connection address is prefilled (ex: `192.168.88.1`). Enter the username and password in the login window. By default, the username is **admin** and there is no password. Select **Connect**.

{{< figure src="mikrotik-routerboard-login.png" alt="MikroTik Routerboard wAP LoRa8 kit Login window" >}}

Once logged in, select the **Gear Box Icon** on the top-right corner of the home page.

{{< figure src="mikrotik-routerboard-home-page.png" alt="MikroTik Routerboard wAP LoRa8 kit Home page window" >}}

Scroll down and click on **LoRa**. In the **LoRa** config window select **Devices**. Select the **LoRa Device** and disable it.

{{< figure src="mikrotik-routerboard-lora-settings.jpeg" alt="MikroTik Routerboard wAP LoRa8 kit LoRa Settings" >}}

{{< figure src="mikrotik-routerboard-lora-disable.jpeg" alt="MikroTik Routerboard wAP LoRa8 kit LoRa disable" >}}

Back in the **LoRa** section, select the **Servers** section. Select the **+** button to add a new server. 

Edit the server parameters.

1. **Name**: A distinct name 
2. **Address**: Address of the Gateway Server. If you followed the [Getting Started guide]({{< ref "/guides/getting-started" >}}) this is the same as what you use instead of `thethings.example.com`.
3. **Up port**: UDP upstream port of the Gateway Server, typically 1700.
4. **Down port**: UDP downstream port of the Gateway Server, typically 1700.

Now back in the **LoRa** section, select **Devices** and select the **LoRa Device**. Click on **Network Servers** and select the server based on the name in the previous step. 

Go back and click on the device and enable it.

{{< figure src="mikrotik-routerboard-lora-enable.png" alt="MikroTik Routerboard wAP LoRa8 kit LoRa enable" >}}

If your configuration was successful, your gateway will connect to {{% tts %}} after a couple of seconds.

## Troubleshooting

If the gateway does not connect to {{% tts %}} after a few minutes, disconnect and reconnect the power supply to power-cycle the gateway. The traffic packets can be observed by going to **LoRa** > **Traffic** to check if the gateway is seeing traffic (Rx/Tx) packets.
