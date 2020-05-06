---
title: Peering
description: ""
summary: Exchange traffic with other LoRaWAN networks via peering to share coverage and improve the overall network performance.
---

## What is it?

Exchange traffic with other LoRaWAN networks via peering to share coverage and improve the overall network performance.

## Who is it for?

Peering is useful for all LoRaWAN public and private network operators. It can improve overall network performance by increasing resilience against gateway failures, expanding coverage area and optimizing end device battery life by communicating with the nearest gateways.

### Typical use cases

1. Forward uplink traffic received by your gateways from devices to their home network. The home network may also use your gateways to transmit downlink traffic to their devices. You can also put commercial agreements in place to monetize coverage.
2. Receive uplink traffic for your devices from other networks. You may also be able to use other networks to send downlink traffic to your devices.

## How does it work?

Your network can be a forwarder and a home network. Forwarder networks have physical gateway infrastructure and home networks have end devices. Most LoRaWAN networks typically have gateways and end devices, so that they can be configured to play both roles. You can also have networks with only gateway infrastructure, configured as forwarder, or only with end devices, configured as home network.

As a forwarder, your network offloads traffic that has been received by your gateways but that is not intended for your network. The offloading goes to a Packet Broker: a LoRaWAN traffic exchange or LoRaWAN roaming hub for {{% tts %}}.

### Packet Broker

[Packet Broker](https://www.packetbroker.org) is a global backbone for LoRaWAN traffic. It is designed to exchange traffic securely between LoRaWAN networks. Packet Broker allows for individual packet selection; networks do not get charged for traffic they did not consume. Packet Broker separates traffic routing from billing and clearing; networks are free to put commercial agreements in place to settle balances. Packet Broker also separates payload from metadata; networks only get charged for the value they need. Finally, {{% tts %}} has native support for Packet Broker and can access the global coverage provided by The Things Network public community network.

Your network authenticates with its NetID and (optionally) a tenant ID to Packet Broker. NetIDs are issued by the LoRa Alliance. To obtain a NetID for your network, [become a member of the LoRa Alliance](https://lora-alliance.org/become-a-member). Alternatively, a host with a large NetID may authorize you as tenant to use Packet Broker.

Packet Broker routes traffic based on device addresses (DevAddr) which are issued from NetIDs: the most significant bits of a 32-bit DevAddr is the NetID. If your network is a tenant of a host NetID, your host configured one or more DevAddr prefixes, and Packet Broker will route traffic to you based on those prefixes.

The Packet Broker Agent of {{% tts %}} connects as forwarder and/or home network to Packet Broker. See the [Enable Peering]({{< relref "enable" >}}) on how to connect your network to Packet Broker.
