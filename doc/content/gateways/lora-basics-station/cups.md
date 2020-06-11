---
title: "Configuration and Update Server Protocol"
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

In your gateway configuration, set the following fields:

CUPS URI: `https://<server-address>:443`

CUPS Key: `<gateway-api-key>`

CUPS Trust: Use the CA certificate of your trust provider
