---
title: "Payload Formatters"
description: ""
---

Payload formatters allow you to process data going to and from end devices. This is useful for converting binary payloads to human readable fields, or for doing any other kind of data conversion on uplinks and downlinks.

This section explains how to set up Application and device specific payload formatters.

<!--more-->

## Application and Device Specific Payload Formatters

Payload formatters can be applied to an entire Application, or to a specific end device. Application payload formatters are useful if all devices use the same binary payload format, or as a fallback when no device specific payload formatter is set.

Device payload formatters allow you to specify a unique payload formatter for each device. Device payload formatters override Application payload formatters.

## Working with Bytes

To work with payload formatters, it is important to understand how payload data is encoded as binary bytes, and how to convert it to meaningful fields.

To see how your device encodes environmental data, see your product datasheet.

See [The Things Network Learn](https://www.thethingsnetwork.org/docs/devices/bytes.html) for an introduction to working with bytes.
