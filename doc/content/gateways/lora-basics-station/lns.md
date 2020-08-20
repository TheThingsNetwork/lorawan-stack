---
title: "LoRaWAN Network Server (LNS)"
description: ""
weight: -1
---

This page contains information about connecting your gateway to {{% tts %}} using the {{% lns %}} (LNS) protocol.

<!--more-->

## Requirements

1. User account on {{% tts %}} with rights to create gateways.

## Create a Gateway

Create a gateway by following the instructions for [Adding Gateways]({{< ref "/gateways/adding-gateways" >}}).

## Create an API Key

LNS requires an API Key with the following rights:
- Link as Gateway to a Gateway Server for traffic exchange, i.e. write uplink and read downlink

If you have not already created one, follow instructions for creating a Gateway API Key in [Adding Gateways]({{< ref "/gateways/adding-gateways" >}}).

## Configure Gateway

On your gateway, set the following configuration fields.

The `<server-address>` is the address of {{% tts %}}. If you followed the [Getting Started guide]({{< ref "/getting-started" >}}) this is the same as what you use instead of `thethings.example.com`.

The `<gateway-api-key>` is the API Key you created above. Create a file named `lns.key` and copy your gateway API Key in as an HTTP header in the following format:

```
Authorization: Bearer <gateway-api-key>
```

Some gateways require that the `lns.key` file is terminated with a Carriage Return Line Feed (`0x0D0A`) character. To easily add a CRLF character and save the file, use the following command:

```bash
echo "Authorization: Bearer <gateway-api-key>" | perl -p -e 's/\r\n|\n|\r/\r\n/g'  > lns.key
```

If using Let's Encrypt to secure your domain, you may download the Let's Encrypt DST X3 Trust file [here](https://letsencrypt.org/certs/lets-encrypt-x3-cross-signed.pem.txt).

TC URI: `wss://<server-address>:8887`

TC Key: `lns.key`

TC Trust: Use the CA certificate of your trust provider
