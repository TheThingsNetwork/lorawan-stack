---
title: "LoRaWAN Network Server Protocol"
description: ""
weight: -1
---

This page contains information about connecting your gateway to {{% tts %}} using the {{% lns %}} (LNS) protocol.

<!--more-->

## Requirements

1. User account on {{% tts %}} with rights to create gateways.

## Create a Gateway

Follow instructions for [Adding Gateways in the Console]({{< ref "/getting-started/console/create-gateway" >}}) or [Adding Gateways Using the Command-line interface]({{< ref "/getting-started/cli/create-gateway" >}}).

## Create an API Key (Optional)

LNS does not require an API Key. If you wish to use Token Authentication, create an API Key with the following rights:
- Link as Gateway to a Gateway Server for traffic exchange, i.e. write uplink and read downlink

If you have not already created one, follow instructions for [Creating a Gateway API Key in the Console]({{< ref "/getting-started/console/create-gateway#create-gateway-api-key" >}}) or [Creating a Gateway API Key Using the Command-line interface]({{< ref "/getting-started/cli/create-gateway#create-gateway-api-key" >}}).

## Configure Gateway

In your gateway configuration, set the following fields:

TC URI: `wss://<server-address>:8887`

TC Key: `<optional-gateway-api-key>`

TC Trust: Use the CA certificate of your trust provider
