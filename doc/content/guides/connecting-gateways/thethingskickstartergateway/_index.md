---
title: "The Things Kickstarter Gateway"
description: ""
weight: 1
---

The Things Kickstarter Gateway is a LoRaWAN gateway, whose technical specifications can be found in [the official documentation](https://www.thethingsnetwork.org/docs/gateways/gateway/). This page guides you to connect it to The Things Stack.

## Prerequisites

1. User account on The Things Stack with rights to create Gateways and API Keys.
2. The Things Gateway running the latest firmware (a minimum of `v1.0.7` is necessary).

## Registration

Create a gateway by following the instructions for the [Console]({{< ref "/guides/getting-started/console#create-gateway" >}}) or the [CLI]({{< ref "/guides/getting-started/cli#create-gateway" >}}). Choose a **Gateway ID** that is at least 6 characters in length. An **EUI** is not necessary.

Create an API Key with Gateway Link rights for this gateway using the same instructions. Copy the key and save it for later use.

## Configuration

Open the front panel of the gateway casing.

While the gateway is powered on, hold the pink reset button for 5 seconds (until each of the 5 LEDs illuminate). This erases the existing configuration on the gateway.

The gateway will now expose a WiFi Access Point whose SSID is of the form `TheThings-Gateway-xxxxx`, to which you should now connect.

In a web browser, open the gateway's configuration page by navigating to http://192.168.84.1/

{{< figure src="ttkg-config-window.png" alt="TTKG Configuration Window" >}}

Enter the following fields:

1. **Name**: the **Gateway ID** that you chose earlier.
2. Choose the WiFi network from the drop down and enter a password if necessary.

Click the **Show Advanced Options** button and enter the following fields:

1. **Account Server**: The URL of The Things Stack. If you're using a port other that `:443` then append that to the URL.
2. **Gateway Key**: The API Key that you created earlier.
3. Click **Save** when done.

This will apply the setting and reboot the gateway. If all the steps have been followed correctly, your gateway will now connect The Things Stack.

## Troubleshooting

If the gateway does not connect to the The Things Stack after a few minutes, disconnect and reconnect the power supply to power-cycle the gateway.
