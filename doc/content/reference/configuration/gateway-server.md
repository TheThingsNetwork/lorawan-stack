---
title: "Gateway Server Options"
description: ""
weight: 3
---

## Forwarding Options

The Gateway Server forwards traffic to upstream hosts based on the `gs.forward` parameter.

- `gs.forward`: Forward the DevAddr prefixes to the specified hosts. This parameter accepts a string in the format `name=devaddrprefixes` (default "=00000000/0")

## Security Options

- `gs.require-registered-gateways`: Require the gateways to be registered in the Identity Server (default "false")

## Basic Station Options

The Gateway Server supports connection of gateways using the Basic Station protocol.

- `gs.basic-station.listen`: Address for the Basic Station frontend to listen on (default ":1887")
- `gs.basic-station.listen-tls`: Address for the Basic Station frontend to listen on (with TLS) (default ":8887")
- `gs.basic-station.use-traffic-tls-address`: Use WSS for the traffic address regardless of the TLS setting (default "false")

The frequency plan to use for unregistered gateways can be set using `gs.basic-station.fallback-frequency-plan-id`. Note that `gs.require-registered-gateways` must be set to false for this to take effect.
- `gs.basic-station.fallback-frequency-plan-id`: Fallback frequency plan ID for non-registered gateways

## MQTT Options

The Gateway Server exposes an MQTT server for connecting gateways via MQTT.

- `gs.mqtt.listen`: Address for the MQTT frontend to listen on (default ":1882")
- `gs.mqtt.listen-tls`: Address for the MQTTS frontend to listen on (default ":8882")
- `gs.mqtt.public-address`: Public address of the MQTT frontend (default "localhost:1882")
- `gs.mqtt.public-tls-address`: Public address of the MQTTs frontend (default "localhost:8882")

## MQTT V2 Options

The Gateway Server exposes an second MQTT server for connecting gateways that use the v2 MQTT format.

- `gs.mqtt-v2.listen`: Address for the MQTT frontend to listen on (default ":1882")
- `gs.mqtt-v2.listen-tls`: Address for the MQTTS frontend to listen on (default ":8882")
- `gs.mqtt-v2.public-address`: Public address of the MQTT frontend (default "localhost:1882")
- `gs.mqtt-v2.public-tls-address`: Public address of the MQTTs frontend (default "localhost:8882")

## UDP Options

The Gateway Server supports the connection of gateways using the Semtech UDP protocol.

- `gs.udp.listeners`: Listen addresses with (optional) fallback frequency plan ID for non-registered gateways. This parameter accepts as string in the format `port=frequency-plan-id`(default ":1700=")
- `gs.udp.schedule-late-time`: Time in advance to send downlink to the gateway when scheduling late (default "800ms")

Options are available to configure connection behavior of UDP gateways.

- `gs.udp.connection-expires`: Time after which a connection of a gateway expires (default "5m0s")
- `gs.udp.downlink-path-expires`: Time after which a downlink path to a gateway expires (default "30s")
- `gs.udp.addr-change-block`: Time to block traffic when a gateway's address changes (default "15s")

Using the `packet-buffer` and `packet-handlers` options, the throughput of UDP packets can be configured.

- `gs.udp.packet-buffer`: Buffer size of unhandled packets (default "50")
- `gs.udp.packet-handlers`: Number of concurrent packet handlers (default "10")

