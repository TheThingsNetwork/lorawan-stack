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

Login to The Things Stack via the CLI/console.

Create a new gateway with the desired **Gateway ID** (at least 6 characters in length) and choose the correct frequency plan for your device. The EUI field can be left blank. For details on using the CLI/Console, refer to the [getting started]({{< ref "/guides/getting-started" >}}) section.

Create an API Key with Gateway Link Rights. Check [here]({{< relref "../../getting-started/console#create-a-gateway-api-key" >}}) for more details. Copy and save the key for later use.

## Configuration

Open the front panel of the gateway casing.

While the gateway is powered on, hold the pink reset button for 5 seconds (until each of the 5 LEDs illuminate). This erases the existing configuration on the gateway.

The gateway will now expose a WiFi Access Point whose SSID is of the form `TheThings-Gateway-xxxxx`, to which you can connect.

In a web browser, navigate to http://192.168.84.1/. A sample page is shown below.
{{< figure src="ttkg-config-window.png" alt="TTKG Configuration Window" >}}

Enter the following fields

1. Name: the **Gateway ID** chosen earlier.
2. Choose the WiFi network from the drop down and enter a password if necessary.

Click the **Show Advanced Options** button and enter the following fields

1. **Account Server**: The URL of The Things Stack. If you're using a port other that `:443` then append that to the URL.
2. **Gateway Key**: The API Key created earlier.
3. Click **Save** when done.

This will apply the setting and reboot the gateway. If all the steps have been followed correctly, your gateway will connect The Things Stack.

## Troubleshooting

If the gateway does not connect to the The Things Stack after a few minutes, disconnect and reconnect the power supply to power-cycle the gateway.
