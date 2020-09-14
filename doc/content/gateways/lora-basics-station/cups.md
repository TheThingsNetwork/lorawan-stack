---
title: "Configuration and Update Server (CUPS)"
description: ""
weight: -1
---

{{% lbs %}} can regularly connect to a {{% cups %}} (CUPS) server to check for configuration and software updates. This page contains information about connecting your gateway to {{% tts %}} to support remote management via the CUPS Protocol.

<!--more-->

> CUPS is **not required** for sending and receiving LoRaWAN data, but it can greatly simplify the management of gateways.

## Requirements

1. User account on {{% tts %}} with rights to create gateways.
2. A gateway which supports {{% lbs %}}.

## Create a Gateway

To connect a gateway using the CUPS protocol, you must first add the gateway in {{% tts %}}. This can be done either in the console, or via the command line. See instructions for [Adding Gateways]({{< ref "/gateways/adding-gateways" >}}). 

## Create an API Key

CUPS requires an API key for your gateway with the following rights:
- View gateway information
- Edit basic gateway settings

To create an API key for your gateway, follow instructions for Creating a Gateway API key in [Adding Gateways]({{< ref "/gateways/adding-gateways" >}}).

## Configure Gateway

Gateway configuration menus differ depending on the manufacturer, but all {{% lbs %}} gateways support the following configuration options. Consult your gateway documentation for more information about configuring your specific gateway. 

### CUPS Server Address

The server address is the network endpoint of {{% tts %}} CUPS. It is a combination of the **protocol** (https), the **server address**, and the **port**:

Enter the following in your gateway as CUPS Server Address: `https://<server-address>:443`

> The `<server-address>` is the address of {{% tts %}}. If you followed the [Getting Started guide]({{< ref "/getting-started" >}}) this is the same as what you use instead of `thethings.example.com`, e.g `https://thethings.example.com:443`

### CUPS Server Certificate / CUPS Trust

This is the [CA certificate](https://en.wikipedia.org/wiki/Certificate_authority) which secures your domain. A `.pem` file containing common certificates is available in the [Root Certificates Reference]({{< ref src="/reference/root-certificates" >}}).

Upload the `.pem` file in your gateway as the CUPS Server Certificate / CUPS Trust.

### CUPS Key File

This is a file which {{% tts %}} uses to verify the identity of your gateway.

Use the following command to create a file called `cups.key` with the `<gateway-api-key>` you created above.

```bash
echo "Authorization: Bearer <gateway-api-key>" | perl -p -e 's/\r\n|\n|\r/\r\n/g'  > cups.key
```

> The above command creates a file called `cups.key`, terminated with a Carriage Return Line Feed (`0x0D0A`) character. Upload this file in your gateway as the CUPS key.
