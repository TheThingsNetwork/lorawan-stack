---
title: "Gateway Configuration Server"
description: ""
weight: 11
---

The Gateway Configuration Server (GCS) generates configuration files for UDP gateways and manages gateway configuration and firmware updates for Basic Station and The Things Kickstarter gateways.

<!--more-->

## Basic Station CUPS

The Gateway Configuration Server implements the [Basic Station CUPS protocol](https://doc.sm.tc/station/cupsproto.html). Gateways implementing this protocol request the GCS for configuration via HTTP(s) POST request. These requests may require authentication in which case the gateway presents the appropriate credentials. The GCS responds to valid requests with a response that contains the URI of Gateway Server to which the gateway should connect.

This response optionally contains signed firmware update data that can be used to remotely update the firmware of supported gateways.

## UDP Configuration File

The Gateway Configuration Server provides an endpoint to query the configuration file for a UDP gateway in JSON format. This allows automated scripts on supported gateways to request the configuration from the server periodically and apply changes in configuration, if applicable, without the need for shell access on the gateway. This configuration file can also be manually queried using API calls.

## The Things Kickstarter Gateway

The Things Kickstarter Gateways connect to the Gateway Configuration Server to fetch key information such as the frequency plan to configure the gateway radio and the MQTT(s) server end point to connect for traffic. Additionally, the firmware of The Things Kickstarter Gateways can be remotely updated by the gateway itself by fetching the necessary files. The location of these files can also be configured per-gateway using the GCS.
