# Gateway Server

The **Gateway Server** component of The Things Network Stack is responsible for the gateways that are connected to it. It includes roughly the same functionality as the **Router** component of our v2 network stack.

## Connectivity to Gateways

Gateways can connect to Gateway Servers over multiple protocols:

+ **UDP**: Gateways can connect to a Gateway Server over [the UDP protocol](https://github.com/Lora-net/packet_forwarder/blob/master/PROTOCOL.TXT).

+ **gRPC**: Gateways can connect to a Gateway Server using [The Things Network's gRPC protocol](../api/gatewayserver.proto#L79).

+ **MQTT**: Though it is not yet implemented, gateways will be able to connect to a Gateway Server using authenticated MQTT.

## Public Gateway Information

## Status Messages

## Forwarding to Network Server

Each gateway server has a list of network servers to connect to.

### Device Address Prefixes



### Peering

## Downlink Scheduling

The gateway server keeps track of all downlinks emitted and to be emitted by gateways connected. The network server can request the scheduling of a downlink with the [`ScheduleDownlink` method](../api/gatewayserver.proto#L89), by attaching the **timestamp** at which a downlink should be sent.

The following checks are being done to schedule a downlink:

+ **Downlink overlap**: If the emission of a downlink overlaps the emission of another downlink, the downlink is dropped.

+ **Time off air**: Some frequency plans have time-off-air constraints, meaning a gateway must not emit for a certain period of time after emitting a downlink.

+ **Duty cycle**: Many countries have **duty cycle restrictions**, prohibiting a device for emitting for more than a certain percentage of time on a certain band. You can find more details in our [official documentation](https://www.thethingsnetwork.org/docs/lorawan/#eu-863-870-mhz-and-duty-cycle).

+ **Dwell time**: Some countries, such as the USA, are subject to **dwell time regulations** - meaning the duration of an transmission can't exceed a certain period.
