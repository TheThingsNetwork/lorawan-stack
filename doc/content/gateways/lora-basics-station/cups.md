---
title: "Configuration and Update Server (CUPS)"
description: ""
weight: -1
---

{{% lbs %}} regularly connects to a {{% cups %}} to check for configuration and software updates. This page contains information about managing your gateway using {{% tts %}} {{% cups %}} (CUPS) Protocol.

<!--more-->

## Requirements

1. User account on {{% tts %}} with rights to create gateways.

## Create a Gateway

Follow instructions for [Adding Gateways]({{< ref "/gateways/adding-gateways" >}}).

## Create an API Key

CUPS requires an API Key with the following rights:
- View gateway information
- Edit basic gateway settings

If you have not already created one, follow instructions for creating a Gateway API Key in [Adding Gateways]({{< ref "/gateways/adding-gateways" >}}).

## Configure Gateway

On your gateway, set the following configuration fields.

The `<server-address>` is the address of {{% tts %}}. If you followed the [Getting Started guide]({{< ref "/getting-started" >}}) this is the same as what you use instead of `thethings.example.com`.

The `<gateway-api-key>` is the API Key you created above. Create a file named `cups.key` and copy your gateway API Key in as an HTTP header in the following format:

```
Authorization: Bearer <gateway-api-key>
```

Some gateways require that the `cups.key` file is terminated with a Carriage Return Line Feed (`0x0D0A`) character. To easily add a CRLF character and save the file, use the following command:

```bash
echo "Authorization: Bearer <gateway-api-key>" | perl -p -e 's/\r\n|\n|\r/\r\n/g'  > cups.key
```

If using Let's Encrypt to secure your domain, you may download the Let's Encrypt DST X3 Trust file [here](https://letsencrypt.org/certs/lets-encrypt-x3-cross-signed.pem.txt).

CUPS URI: `https://<server-address>:443`

CUPS Key: `cups.key`

CUPS Trust: Use the CA certificate of your trust provider
