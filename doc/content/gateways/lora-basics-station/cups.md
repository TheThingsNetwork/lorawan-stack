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

Follow instructions for [Adding Gateways in the Console]({{< ref "/getting-started/console/create-gateway" >}}) or [Adding Gateways Using the Command-line interface]({{< ref "/getting-started/cli/create-gateway" >}}).

## Create an API Key

CUPS requires an API Key with the following rights:
- View gateway information
- Edit basic gateway settings

If you have not already created one, follow instructions for [Creating a Gateway API Key in the Console]({{< ref "/getting-started/console/create-gateway#create-gateway-api-key" >}}) or [Creating a Gateway API Key Using the Command-line interface]({{< ref "/getting-started/cli/create-gateway#create-gateway-api-key" >}}).

## Configure Gateway

On your gateway, set the following configuration fields.

The `<server-address>` is the address of {{% tts %}}. If you followed the [Getting Started guide]({{< ref "/getting-started" >}}) this is the same as what you use instead of `thethings.example.com`.

The `<gateway-api-key>` is the API Key you created above. Copy your gateway API Key in to a `gateway-api.key` file (the filename is not important) as an HTTP header in the following format:

```
Authorization: <gateway-api-key>
```

If using Let's Encrypt to secure your domain, you may download the Let's Encrypt DST X3 Trust file [here](https://letsencrypt.org/certs/lets-encrypt-x3-cross-signed.pem.txt).

CUPS URI: `https://<server-address>:443`

CUPS Key: `<gateway-api-key>`

CUPS Trust: Use the CA certificate of your trust provider
