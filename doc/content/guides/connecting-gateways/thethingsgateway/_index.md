---
title: "The Things Gateway"
description: ""
weight: 2
menu:
  main:
    weight: 2
---

The Things Gateway is an MQTT based single channel LoRaWAN gateway. The technical specifications can be found in [the official documentation](https://www.thethingsnetwork.org/docs/gateways/gateway/). This page guides you to connect it to a The Things Stack instance.


## Preparation

* User account on a The Things Stack instance with rights to create Gateways and API Keys.
* The Things Gateway running the latest firmware (a minimum of `v1.0.7` is necessary).

## Registration

* Login to your The Things Stack instance via the CLI/console.
* Create a new gateway with the desired `Gateway ID` and choose the correct frequency plan for your device. The EUI field can be left blank.
  * For details on using the CLI/Console, refer to the [getting started](../../getting-started/) section.
* Create an API Key with Gateway Link Rights (TODO: needs more info?). Copy and save the key for later use.

## Configuration

* Open the front panel of the gateway casing.
* While the gateway is powered on, hold the pink reset button for 5 seconds (until each of the 5 LEDs illuminate).
  * This erases the existing configuration on the gateway.
* The gateway will now expose a WiFi Access Point whose SSID is of the form `TheThings-Gateway-xxxxx`.
  * Connect to this network using a mobile phone/computer.
* In a web browser, navigate to http://192.168.84.1/. A sample page is shown below.
{{< figure src="ttg-config-window.png" alt="TTG Configuration Window" >}}
* Enter the following fields
  * Name: the `Gateway ID` chosen earlier.
  * Choose the WiFi network from the drop down and enter a password if necessary.
* Click the `Show Advanced Options` button and enter the following fields
  * Account Server: The URL of your stack's identity server. This is of the form `https://<your-domain-or-ip>:<port>`.
  * (TODO: How to make this URL generic?)
  * Gateway Key: The API Key created earlier.
* Click `Save` when done.
* This will apply the setting and reboot the gateway. If all have been followed correctly, your gateway will connect to your stack instance.

## Troubleshooting

* If the gateway doesn't connect to the backend after a few minutes, disconnect and reconnect the power supply to power-cycle it.
