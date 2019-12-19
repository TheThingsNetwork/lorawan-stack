---
title: "Gateway Configuration Server Options"
description: ""
weight: 9
---

## Security Options

- `gcs.require-auth`: Require authentication for the HTTP endpoints

## Basic Station CUPS Options

The `gcs.basic-station` options configure the GCS to handle Basic Station CUPS requests.

- `gcs.basic-station.allow-cups-uri-update`: Allow CUPS URI updates
- `gcs.basic-station.default.lns-uri`: The default LNS URI that the gateways should use. If no Gateway Server address is registered, the default value is used.
- `gcs.basic-station.owner-for-unknown.account-type`: Type of account to register unknown gateways to (user|organization)
- `gcs.basic-station.owner-for-unknown.api-key`: API Key to use for unknown gateway registration
- `gcs.basic-station.owner-for-unknown.id`: ID of the account to register unknown gateways to
- `gcs.basic-station.require-explicit-enable`: Require gateways to explicitly enable CUPS

## The Things Kickstarter Gateway Options

The `gcs.the-things-gateway.firmware-url` and `gcs.the-things-gateway.update-channel` options configure the source of firmware updates for The Things Kickstarter Gateway.

- `gcs.the-things-gateway.default.firmware-url`: The default URL to the firmware storage
- `gcs.the-things-gateway.default.mqtt-server`: The default MQTT server that the gateways should use. If no Gateway Server address is registered, the default value is used. The format is `mqtts://<IP-or-Address>:port`.
- `gcs.the-things-gateway.default.update-channel`: The default update channel that the gateways should use
