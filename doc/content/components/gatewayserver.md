---
title: "Gateway Server"
weight: 4
--- 

# Gateway Server

The **Gateway Server** component of The Things Network Stack is responsible for the gateways that are connected to it. It includes roughly the same functionality as the **Router** component of our v2 network stack.

## Connectivity to Gateways

Gateways can connect to Gateway Servers over multiple protocols.

#### UDP protocol

Gateways can connect to a Gateway Server over [the UDP protocol](https://github.com/Lora-net/packet_forwarder/blob/master/PROTOCOL.TXT). The EUI that is sent with every message is used to identify the gateway.

If a gateway is found in the Identity Server with this EUI, messages are correlated to this gateway. Otherwise, uplinks are still routed. However, the gateway will not send downlinks to this gateway, given that its regional parameters cannot be identified.

Many packet forwarders implementing this protocol do not implement any queuing system for downlinks, resulting in packet loss since SX1301 concentrators cannot buffer multiple downlinks. The Things Network thus implements, for the UDP protocol, a delay to sent downlinks to gateway just before they're meant to be emitted by the concentrator. You can disable this feature individually per gateway, for example if the RTT between your gateway and the gateway server is too high.

#### gRPC protocol

Gateways can connect to a Gateway Server using [The Things Network's gRPC protocol](../api/gatewayserver.proto).

#### MQTT protocol

Gateways can connect to a Gateway Server by exchanging [protobuf-encoded messages](../api/gatewayserver.proto) over MQTT. This preserves the security and bandwidth usage features of the gRPC protocol, and is compatible with environments for which gRPC is not available.

Clients should:

+ Set the broker address to the gateway server's [MQTT endpoint](networking.md).

+ Set the ID of the gateway as username, and an API key with the `GATEWAY_INFO` and `GATEWAY_LINK` rights as password.

+ Use the `v3/<gateway ID>/up` topic to push uplinks, and the `v3/<gateway ID>/status` to push status messages.

+ Subscribe to the `v3/<gateway ID>/down` topic to pull downlinks.

## Gateway Information

While a gateway is connected, the Gateway Server collects statistics about the messages exchanged with the gateway, and about the status messages sent by the gateway. Those statistics can be retrieved using the `GetGatewayObservations` endpoint.

## Communication with Network Server

The main function of the Gateway Server is to establish a stable connection with gateways, and to serve as a relay between those gateways and the Network Servers.

When routing an uplink, the Gateway Server decides which Network Server to send it to based on the `DevAddr` of the device. Join request are similarly routed based on the `DevEUI` of the device.

The Gateway Server exchanges with Network Servers over gRPC. It claims identifiers of gateways using `ClaimIDs` and `UnclaimIDs` in the cluster, and sends uplinks using `HandleUplink`. It exposes an service with `ScheduleDownlink`, that the Network Server uses to send downlinks to gateways.

## Downlink Scheduling

The gateway server keeps track of all downlinks emitted and to be emitted by gateways connected. The network server can request the scheduling of a downlink with the [`ScheduleDownlink` method](../api/gatewayserver.proto), by attaching the **timestamp** at which a downlink should be sent. The gateway server does not decide which gateway sends a downlink, as this is decided by the network server.

The following checks are being done by the gateway server before scheduling a downlink:

+ **Downlink overlap**: If the emission of a downlink overlaps the emission of another downlink, the downlink is dropped.

+ **Time off air**: Some frequency plans have time-off-air constraints, meaning a gateway must not emit for a certain period of time after emitting a downlink.

+ **Duty cycle**: Many countries have **duty cycle restrictions**, prohibiting a device for emitting for more than a certain percentage of time on a certain band. You can find more details in our [official documentation](https://www.thethingsnetwork.org/docs/lorawan/#eu-863-870-mhz-and-duty-cycle).

+ **Dwell time**: Some countries, such as the United States, are subject to **dwell time regulations** - meaning the duration of an transmission can't exceed a certain period.
