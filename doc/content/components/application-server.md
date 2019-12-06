---
title: "Application Server"
description: ""
weight: 4
---

The Application Server handles the LoRaWAN application layer, including uplink data decryption and decoding, downlink queuing and downlink data encoding and encryption.

It hosts an MQTT server for streaming application data, supports HTTP webhooks as well as pub/sub integrations.

<!--more-->

## Linking to Network Servers

Application Servers link to Network Servers to receive upstream traffic and write downstream traffic.

Most {{% tts %}} clusters contain an Application Server, but you can also link an external Application Server to a Network Server. This ensures that the application session key (AppSKey) is not available to the network-layer for end-to-end security.

Only one Application Server instance can be linked to a Network Server at a time.

## Connectivity

Applications can connect to Application Server over multiple protocols and mechanisms.

### MQTT Protocol

Applications can connect to an Application Server by exchanging JSON messages over MQTT. MQTT is available over TLS, providing confidentiality of messages exchanged between applications and the Application Server.

The upstream messages do not only contain data uplink messages, but also join-accepts and downlink events, on separate topics.

See [Application Server MQTT server]({{< ref "/reference/application-server-data/mqtt" >}}) for more information.

### HTTP Webhooks

Applications can get streaming JSON messages via HTTP webhooks, and schedule downlink messages by making an HTTP request to the Application Server.

Like MQTT, all upstream messages can be configured, including uplink messages, join-accepts and downlink events, each to separate URL paths.

See [HTTP webhooks]({{< ref "/reference/application-server-data/webhooks" >}}) for more information.

### Pub/Sub Integrations

Applications can also use pub/sub integrations to work with streaming data. This includes connecting to an external MQTT server and [NATS server](https://www.nats.io).

## Message Processing

The Application Server can decode and encode binary payload sent and received by end devices. This allows for working with structured streaming data, such as JSON objects using MQTT and HTTP webhooks, yet using compact binary data that is transmitted over the air.

Message processors can be well-known formats or custom scripts, and can be set on the device level, or for the entire application.
