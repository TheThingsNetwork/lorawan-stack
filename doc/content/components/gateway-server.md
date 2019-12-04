---
title: "Gateway Server"
description: ""
weight: 2
---

The Gateway Server maintains connections with gateways supporting the Basic Station, UDP, MQTT and gRPC protocols. It forwards uplink traffic to Network Servers directly or indirectly, and schedules downlink traffic on gateways.

<!--more-->

## Connectivity

Gateways can connect to Gateway Servers over multiple protocols.

### Basic Station Protocol

Gateways can connect to a Gateway Server using the [Basic Station](https://doc.sm.tc/station/index.html) LNS protocol. This is the recommended protocol for connecting gateways.

### UDP Protocol

Gateways can connect to a Gateway Server over [the UDP protocol](https://github.com/Lora-net/packet_forwarder/blob/master/PROTOCOL.TXT). The EUI that is sent with every message is used to identify the gateway.

If a gateway is found in the Identity Server with this EUI, messages are correlated to this gateway. Otherwise, depending on network configuration, uplinks may routed or dropped. However, the network will not send downlinks to this gateway, given that its regional parameters cannot be identified.

Older versions of packet forwarders implementing the UDP protocol do not implement any queuing system for downlinks, resulting in packet loss since SX130x concentrators do not buffer multiple downlinks. {{% tts %}} therefore implements, for the UDP protocol, a delay to send downlinks to gateway just before they're meant to be emitted by the concentrator. You can disable this feature individually per gateway, for example if the RTT between your gateway and the Gateway Server is too high.

### MQTT Protocol

Gateways can connect to a Gateway Server by exchanging [protocol buffers](https://developers.google.com/protocol-buffers) over MQTT. MQTT is available over TLS, providing confidentiality of messages exchanged between the gateway and the network. The encoding with protocol buffers reduces bandwidth usage compared to the UDP protocol, which uses JSON encoding.

Packet forwarders implementing the MQTT protocols are specific for {{% tts %}}.

## Gateway Information

While a gateway is connected, the Gateway Server collects statistics about the messages exchanged with the gateway, and about the status messages sent by the gateway. Those statistics can be retrieved from the Gateway Server using its gRPC and HTTP APIs. See [`Gs` service]({{< ref "/reference/api/gateway_server#Gs" >}}).

## Communication with Network Server

The main function of the Gateway Server is to maintain connections with gateways, and to serve as a relay between those gateways and the Network Servers.

### Uplink Messages

When receiving a data uplink message, the Gateway Server decides which Network Server to send it to based on the `DevAddr` of the device and the configured forwarding table with endpoints. Join-requests are routed to all configured endpoints. An endpoint can be the cluster's Network Server over gRPC, or an intermediate traffic routing mechanism.

### Downlink Messages

Network Servers can request transmission for downlink messages. The Gateway Server attempts to schedule the message based on the selected gateways, time to send the message and LoRaWAN settings (downlink class, RX1 delay and RX1/RX2 data rates and frequencies).

The Gateway Server keeps track of all downlinks emitted and to be emitted by gateways connected to it, including the exact time-on-air based on message size and data rate. This allows {{% tts %}} to do smart scheduling. Besides timing and LoRaWAN settings, the Gateway Server takes applicable limitations into account, including:

- **Scheduling conflicts**: If the emission of a downlink overlaps the emission of another downlink, the downlink is refused.
- **Time-off-air**: Some bands and frequency plans have time-off-air constraints, meaning a gateway must not emit for a certain period of time after emitting a downlink.
- **Duty-cycle**: Some countries, such as European countries, have **duty-cycle restrictions**, prohibiting a device for emitting for more than a certain percentage of time on a certain band.
- **Dwell time**: Some countries, such as the United States, are subject to **dwell time regulations**, meaning the duration of a transmission cannot exceed a certain limit.
