---
title: "LoRaWAN Network Server (LNS)"
description: ""
weight: -1
---

LNS establishes a data connection between a {{% lbs %}} and {{% tts %}}. This page contains information about connecting your gateway to {{% tts %}} using the {{% lns %}} (LNS) protocol.

<!--more-->

> The LNS protocol is **required** for sending and receiving LoRaWAN data with {{% lbs %}}, while the CUPS protocol is not.

## Requirements

1. User account on {{% tts %}} with rights to create gateways.
2. A gateway which support {{% lbs %}}.

## Create a Gateway

To connect a gateway using the LNS protocol, you must first add the gateway in {{% tts %}}. This can be done either in the console, or via the command line. See instructions for [Adding Gateways]({{< ref "/gateways/adding-gateways" >}}). 

## Create an API Key

LNS requires an API Key with the following rights:
- Link as Gateway to a Gateway Server for traffic exchange, i.e. write uplink and read downlink

To create an API key for your gateway, follow instructions for Creating a Gateway API key in [Adding Gateways]({{< ref "/gateways/adding-gateways" >}}).

## Configure Gateway

Gateway configuration menus differ depending on the manufacturer, but all {{% lbs %}} gateways support the following configuration options. Consult your gateway documentation for more information about configuring your specific gateway. 

### LNS Server Address

The server address is the network endpoint of {{% tts %}} LNS. It is a combination of the **protocol** (wss), the **server address**, and the **port**:

Enter the following in your gateway as the LNS Server Address: `wss://<server-address>:8887`

> The `<server-address>` is the address of {{% tts %}}. If you followed the [Getting Started guide]({{< ref "/getting-started" >}}) this is the same as what you use instead of `thethings.example.com`, e.g `wss://thethings.example.com:8887`

### LNS Server Certificate / LNS Trust

This is the [CA certificate](https://en.wikipedia.org/wiki/Certificate_authority) which secures your domain. A `.pem` file containing common certificates is available in the [Root Certificates Reference]({{< ref "/reference/root-certificates" >}}).

Upload the `.pem` file in your gateway as the LNS Server Certificate / LNS Trust.

### LNS Key File

This is a file which {{% tts %}} uses to verify the identity of your gateway.

Use the following command to create a file called `lns.key` with the `<gateway-api-key>` you created above.

```bash
echo "Authorization: Bearer <gateway-api-key>" | perl -p -e 's/\r\n|\n|\r/\r\n/g'  > lns.key
```

> The above command creates a file called `lns.key`, terminated with a Carriage Return Line Feed (`0x0D0A`) character. Upload this file in your gateway as the LNS key.
