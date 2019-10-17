---
title: "Gateway Server"
description: ""
weight: 2
---

The Gateway Server maintains connections with gateways supporting the UDP, MQTT, gRPC and Basic Station protocols. It forwards uplink traffic to Network Servers, and schedules downlink traffic on gateways.

<!--more-->

## Connectivity

Gateways can connect to Gateway Servers over multiple protocols.

### Basic Station protocol

Gateways can connect to a Gateway Server using the [Basic Station](https://doc.sm.tc/station/index.html) LNS protocol.

### UDP protocol

Gateways can connect to a Gateway Server over [the UDP protocol](https://github.com/Lora-net/packet_forwarder/blob/master/PROTOCOL.TXT). The EUI that is sent with every message is used to identify the gateway.

If a gateway is found in the Identity Server with this EUI, messages are correlated to this gateway. Otherwise, depending on network configuration, uplinks may routed or dropped. However, the network will not send downlinks to this gateway, given that its regional parameters cannot be identified.

Many packet forwarders implementing the UDP protocol do not implement any queuing system for downlinks, resulting in packet loss since SX1301 concentrators cannot buffer multiple downlinks. The Things Stack therefore implements, for the UDP protocol, a delay to send downlinks to gateway just before they're meant to be emitted by the concentrator. You can disable this feature individually per gateway, for example if the RTT between your gateway and the Gateway Server is too high.

### MQTT protocol

Gateways can connect to a Gateway Server by exchanging [protocol buffers](https://developers.google.com/protocol-buffers) over MQTT. MQTT is also available over TLS, providing confidentiality of messages exchanged between the gateway and the network. The encoding with protocol buffers reduces bandwidth usage compared to the UDP protocol, which uses JSON encoding.

## Gateway Information

While a gateway is connected, the Gateway Server collects statistics about the messages exchanged with the gateway, and about the status messages sent by the gateway. Those statistics can be retrieved from the Gateway Server using its gRPC and HTTP APIs.

## Communication with Network Server

The main function of the Gateway Server is to establish a stable connection with gateways, and to serve as a relay between those gateways and the Network Servers.

When routing an uplink, the Gateway Server decides which Network Server to send it to based on the `DevAddr` of the device. Join request are similarly routed based on the `DevEUI` of the device.

The Gateway Server exchanges with Network Servers over gRPC. It claims identifiers of gateways using `ClaimIDs` and `UnclaimIDs` in the cluster, and sends uplinks using `HandleUplink`. It exposes an service with `ScheduleDownlink`, that the Network Server uses to send downlinks to gateways.

## Downlink Scheduling

The Gateway Server keeps track of all downlinks emitted and to be emitted by gateways connected to it. This allows the stack to do smarter scheduling. When the Network Server requests the Gateway Server to schedule a downlink message on a gateway, it checks the schedule to see if a downlink is possible:

- **Downlink overlap**: If the emission of a downlink overlaps the emission of another downlink, the downlink is refused.
- **Time off air**: Some frequency plans have time-off-air constraints, meaning a gateway must not emit for a certain period of time after emitting a downlink.
- **Duty cycle**: Many countries have **duty cycle restrictions**, prohibiting a device for emitting for more than a certain percentage of time on a certain band.
- **Dwell time**: Some countries, such as the United States, are subject to **dwell time regulations** - meaning the duration of a transmission can't exceed a certain limit.
