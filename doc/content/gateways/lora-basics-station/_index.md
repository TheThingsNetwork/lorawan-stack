---
title: "LoRa Basics™ Station"
description: ""
weight: -1
---

The [{{% lbs %}}](https://lora-developers.semtech.com/resources/tools/basic-station/welcome-basic-station/) protocol simplifies management of large scale LoRaWAN networks. {{% lbs %}} is the preferred way of connecting Gateways to {{% tts %}}.

This section contains information for connecting your gateway to {{% tts %}} using {{% lbs %}} and its sub protocols.

<!--more-->

## Advantages of {{% lbs %}}

Some of the advantages of {{% lbs %}} over the legacy UDP Packet Forwarder are:

- Centralized Update and Configuration Management
- TLS and Token-based Authentication
- Centralized Channel-Plan Management
- No Dependency on Local Time Keeping

## LNS and CUPS Sub Protocols

{{% lbs %}} contains two sub protocols for connecting Gateways to Network Servers, LoRaWAN Network Server (LNS) and Configuration and Update Server (CUPS).


### LoRaWAN Network Server (LNS)

LNS establishes a data connection between a {{% lbs %}} gateway and a Network Server (in this case, {{% tts %}}). LoRa® uplink and downlink frames are exchanged through this data connection. The LNS protocol is **required** for sending and receiving LoRaWAN data.

### Configuration and Update Server (CUPS)

CUPS allows a Network Server to configure gateways remotely, and to update gateway firmware. CUPS is **not required** for sending and receiving LoRaWAN data, but it can greatly simplify the management of gateways.

More information about {{% lbs %}} is available at [Semtech's Developer Portal](https://lora-developers.semtech.com/resources/tools/basic-station/welcome-basic-station/)
