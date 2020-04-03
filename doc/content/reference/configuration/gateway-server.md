---
title: "Gateway Server Options"
description: ""
---

## Forwarding Options

The Gateway Server forwards traffic to upstream hosts based on the `gs.forward` parameter.

- `gs.forward`: Forward the DevAddr prefixes to the specified hosts. This parameter accepts a list of strings in the format `target=dev-addr-prefix1,dev-addr-prefix2`

Enter your NetID to view recommended forwarding settings:

{{< dev-addr-prefix >}}

- <code>"cluster=00000000/0"</code>: Forward all traffic to the Network Server in the cluster (default)
- <code>"cluster=<span data-content="dev-addr-prefix"></span>"</code>: Forward only traffic for your NetID to the Network Server
- <code>"cluster=00000000/6,<span data-content="dev-addr-prefix"></span>"</code>: Forward only traffic for the experimental NetIDs `000000` and `000001`, and your NetID to the Network Server in the cluster
- <code>"cluster=00000000/6,<span data-content="dev-addr-prefix"></span> packetbroker=00000000/0"</code>: Forward only traffic for the experimental NetIDs `000000` and `000001`, and your NetID to the Network Server in the cluster, and all traffic to Packet Broker (see [configuration]({{< relref "packet-broker-agent.md" >}}))

## Location Update Options

The Gateway Server can be configured to update the location of gateway antennas from incoming status messages.

- `gs.update-gateway-location-debounce-time`: Debounce time for gateway location updates from status messages.

## Security Options

- `gs.require-registered-gateways`: Require the gateways to be registered in the Identity Server

## Basic Station Options

The Gateway Server supports connection of gateways using the Basic Station protocol.

- `gs.basic-station.listen`: Address for the Basic Station frontend to listen on
- `gs.basic-station.listen-tls`: Address for the Basic Station frontend to listen on (with TLS)
- `gs.basic-station.use-traffic-tls-address`: Use WSS for the traffic address regardless of the TLS setting

The frequency plan to use for unregistered gateways can be set using `gs.basic-station.fallback-frequency-plan-id`. Note that `gs.require-registered-gateways` must be set to false for this to take effect.

- `gs.basic-station.fallback-frequency-plan-id`: Fallback frequency plan ID for non-registered gateways

## MQTT Options

The Gateway Server exposes an MQTT server for connecting gateways via MQTT.

- `gs.mqtt.listen`: Address for the MQTT frontend to listen on
- `gs.mqtt.listen-tls`: Address for the MQTTS frontend to listen on
- `gs.mqtt.public-address`: Public address of the MQTT frontend
- `gs.mqtt.public-tls-address`: Public address of the MQTTs frontend

## MQTT V2 Options

The Gateway Server exposes an second MQTT server for connecting gateways that use the V2 MQTT format.

- `gs.mqtt-v2.listen`: Address for the MQTT frontend to listen on
- `gs.mqtt-v2.listen-tls`: Address for the MQTTS frontend to listen on
- `gs.mqtt-v2.public-address`: Public address of the MQTT frontend
- `gs.mqtt-v2.public-tls-address`: Public address of the MQTTs frontend

## UDP Options

The Gateway Server supports the connection of gateways using the Semtech UDP protocol.

- `gs.udp.listeners`: Listen addresses with (optional) fallback frequency plan ID for non-registered gateways. This parameter accepts as string in the format `listen-address=frequency-plan-id`
- `gs.udp.schedule-late-time`: Time in advance to send downlink to the gateway when scheduling late

Options are available to configure connection behavior of UDP gateways.

- `gs.udp.connection-expires`: Time after which a connection of a gateway expires
- `gs.udp.downlink-path-expires`: Time after which a downlink path to a gateway expires
- `gs.udp.addr-change-block`: Time to block traffic when a gateway's address changes

Using the `packet-buffer` and `packet-handlers` options, the throughput of UDP packets can be configured.

- `gs.udp.packet-buffer`: Buffer size of unhandled packets
- `gs.udp.packet-handlers`: Number of concurrent packet handlers

## Statistics Options

Specify options for gateway connection statistics:

- `gs.update-connection-stats-debounce-time`: Time before repeated refresh of the gateway connection stats

## UDP Rate Limiting Options

The Gateway Server supports rate limiting traffic for gateways that are using the Semtech UDP protocol.

- `gs.udp.rate-limiting.enable`: Enable rate limiting for gateways
- `gs.udp.rate-limiting.messages`: Number of past messages to check timestamp for
- `gs.udp.rate-limiting.threshold`: Filter packet if timestamp is not newer than the older timestamps of the previous messages by this threshold
