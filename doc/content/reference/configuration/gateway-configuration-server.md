---
title: "Gateway Configuration Server Options"
description: ""
weight: 9
---

## Security Options

- `gcs.require-auth`: Require authentication for the HTTP endpoints

## Basic Station CUPS Options

The `gcs.basic-station` options configure the GCS to handle Basic Station CUPS requests.

- `gcs.basic-station.allow-cups-uri-update`: Allow CUPS URI updates (default "false")
- `gcs.basic-station.default.lns-uri`: The default LNS URI that the gateways should use. If no Gateway Server address is registered, the default value is used (default "wss://localhost:8887")
- `gcs.basic-station.owner-for-unknown.account-type`: Type of account to register unknown gateways to (user|organization)
- `gcs.basic-station.owner-for-unknown.api-key`: API Key to use for unknown gateway registration
- `gcs.basic-station.owner-for-unknown.id`: ID of the account to register unknown gateways to
- `gcs.basic-station.require-explicit-enable`: Require gateways to explicitly enable CUPS (default "false")

## The Things Kickstarter Gateway Options

The `gcs.the-things-gateway.firmware-url` and `gcs.the-things-gateway.updated-channel` options configure the source of firmware updates for The Things Kickstarter Gateway.

- `gcs.the-things-gateway.default.firmware-url`: The default URL to the firmware storage (default "https://thethingsproducts.blob.core.windows.net/the-things-gateway/v1")
- `gcs.the-things-gateway.default.mqtt-server`: The default MQTT server that the gateways should use. If no Gateway Server address is registered, the default value is used (default "mqtts://localhost:8881")
- `gcs.the-things-gateway.default.update-channel`: The default update channel that the gateways should use (default "stable")
