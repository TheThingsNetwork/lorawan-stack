---
title: "Components"
description: ""
weight: 4
menu:
  main:
    weight: 4
---

TTN Stack is composed of components. The core components Network Server, Application Server and Join Server follow the LoRaWAN Network Reference Model. TTN Stack also contains an Identity Server, Gateway Server and Console.

{{< figure src="components.png" alt="TTN Stack components" width="80%" height="80%" >}}

- **Identity Server**: stores applications, end devices, gateways, users, organizations, OAuth clients, API keys and collaborators. Also acts as a OAuth 2.0 server with login and consent screens
- **Gateway Server**: maintains connections with gateways supporting the UDP, MQTT, gRPC and Basic Station protocols, forwards uplink traffic to Network Servers, schedules downlink traffic
- **Network Server**: handles LoRaWAN network layer, including MAC commands, regional parameters and adaptive data rate (ADR)
- **Application Server**: handles LoRaWAN application layer, including uplink data decryption and decoding, downlink data encoding and encryption, downlink queuing and hosts an MQTT server for streaming application data, manages HTTP webhooks and pub/sub integrations
- **Join Server**: handles LoRaWAN join flow, including Network and Application Server authentication and session key generation
- **Console**: provides a web interface for managing the components
- **Command-line Interface**: provides a cross-platform interface for managing components through command-line
- Supporting components
  - **Gateway Configuration Server**: generates configuration files for gateways and manages gateway firmware updates
  - **Device Template Converter**: converts data to devices templates for migrating networks and importing vendor-specific data
  - **Device Repository** (in progress): stores devices brands and models, their capabilities, payload formatters and device templates
