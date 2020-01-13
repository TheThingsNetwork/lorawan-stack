---
title: "Application Server Options"
description: ""
weight: 5
---

## Linking Options

The Application Server links to a Network Server. The `link-mode` configures how linking occurs.

- `as.link-mode`: Mode to link applications to their Network Server (all, explicit) (default "all")

## Security Options

- `as.device-kek-label`: Label of KEK used to encrypt device keys at rest

## Interoperability Options

The `as.interop` options configure how Application Server performs interoperability with other LoRaWAN Backend Interfaces-compliant servers.

- `as.interop.id`: AS-ID used for interoperability
- `as.interop.config-source`: Source of the interoperability client configuration (directory, url, blob)
- `as.interop.blob.bucket`: Blob bucket, which contains interoperability client configuration
- `as.interop.blob.path`: Blob path, which contains interoperability client configuration
- `as.interop.directory`: OS filesystem directory, which contains interoperability client configuration
- `as.interop.url`: URL, which contains interoperability client configuration

## MQTT Options

Application Server exposes an MQTT server for streaming data.

- `as.mqtt.listen`: Address for the MQTT frontend to listen on (default ":1883")
- `as.mqtt.listen-tls`: Address for the MQTTS frontend to listen on (default ":8883")
- `as.mqtt.public-address`: Public address of the MQTT frontend (default "localhost:1883")
- `as.mqtt.public-tls-address`: Public address of the MQTTs frontend (default "localhost:8883")

## HTTP Webhooks Options

Application Server has an internal queue with worker routines for outgoing requests. When remote endpoints are not fast enough and queue (with `queue-size`) gets full, new traffic gets discarded. You can tune these parameters for optimal performance, considering memory consumption with a large queue size and number of workers.

- `as.webhooks.queue-size`: Number of requests to queue (default 16)
- `as.webhooks.target`: Target of the integration (direct) (default "direct")
- `as.webhooks.timeout`: Wait timeout of the target to process the request (default 5s)
- `as.webhooks.workers`: Number of workers to process requests (default 16)

Application Server supports templates for webhooks that can be loaded from a `directory` or `url`.

- `as.webhooks.templates.directory`: Retrieve the webhook templates from the filesystem
- `as.webhooks.templates.url`: Retrieve the webhook templates from a web server
- `as.webhooks.templates.logo-base-url`: The base URL for the logo storage

Application Server supports communicating the paths of the downlink queue operations to the webhook endpoints via headers. The paths are computed from the public address, and the HTTPS endpoint is preferred over the HTTP one.

- `as.webhooks.downlinks.public-address`: Public address of the HTTP webhooks frontend (default "http://localhost:1885/api/v3")
- `as.webhooks.downlinks.public-tls-address`: Public address of the HTTPS webhooks frontend
