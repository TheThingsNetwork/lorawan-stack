---
title: "Working with Events"
description: ""
weight: 20
---

{{% tts %}} generates lots of events that allow you to get insight in what is going on. You can subscribe to application, gateway, end device events, as well as to user, organization and OAuth client events.

This guide shows how to subscribe to events with the CLI and with the HTTP API.

<!--more-->

## Subscribe with CLI

To follow your gateway `gtw1` and application `app1` events at the same time:

```bash
$ ttn-lw-cli events subscribe --gateway-id gtw1 --application-id app1
```

## Subscribe with HTTP

You can get streaming events with `curl`. For this, you need an API key for the entities you want to watch, for example:

```bash
$ ttn-lw-cli user api-key create \
  --user-id admin \
  --right-application-all \
  --right-gateway-all
```

With the created API key:

```bash
$ curl https://thethings.example.com/api/v3/events \
  -X POST \
  -H "Authorization: Bearer NNSXS.BR55PTYILPPVXY.." \
  --data '{"identifiers":[{"application_ids":{"application_id":"app1"}},{"gateway_ids":{"gateway_id":"gtw1"}}]}'
```

>Note: The created API key for events is highly privileged; do not use it if you don't need it for events.

## Example: join flow

These are the events of a typical join flow:

```js
{
  "name": "gs.up.receive", // Gateway Server received an uplink message from a device.
  "time": "2019-04-04T09:54:34.786220Z",
  "identifiers": [
    {
      "gateway_ids": {
        "gateway_id": "multitech",
        "eui": "00800000A0000DB4"
      }
    }
  ],
  "correlation_ids": [
    "gs:conn:01D7KWADW2E5CJA32VS1MTR2J6",
    "gs:uplink:01D7KWB0N2KVCV8HZABC8DDHSA"
  ]
}
{
  "name": "js.join.accept", // Join Server accepted the join-accept.
  "time": "2019-04-04T09:54:34.806812Z",
  "identifiers": [
    {
      "device_ids": {
        "device_id": "dev1",
        "application_ids": {
          "application_id": "app1"
        },
        "dev_eui": "4200000000000000",
        "join_eui": "4200000000000000"
      }
    }
  ],
  "correlation_ids": [
    "rpc:/ttn.lorawan.v3.NsJs/HandleJoin:01D7KWB0NCTDY835V5N3CYWBZK"
  ]
}
{
  "name": "ns.up.join.forward", // Network Server forwarded the join-accept and it got accepted.
  "time": "2019-04-04T09:54:34.808132Z",
  "identifiers": [
    {
      "device_ids": {
        "device_id": "dev1",
        "application_ids": {
          "application_id": "app1"
        },
        "dev_eui": "4200000000000000",
        "join_eui": "4200000000000000"
      }
    }
  ],
  "correlation_ids": [
    "gs:conn:01D7KWADW2E5CJA32VS1MTR2J6",
    "gs:uplink:01D7KWB0N2KVCV8HZABC8DDHSA",
    "ns:uplink:01D7KWB0N5C1T8TE2HAVBJN5Y4",
    "rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D7KWB0N5G2N5C0AFXT4YMF8R"
  ]
}
{
  "name": "ns.up.merge_metadata", // Network Server merged metadata of incoming uplink messages.
  "time": "2019-04-04T09:54:34.991332Z",
  "identifiers": [
    {
      "device_ids": {
        "device_id": "dev1",
        "application_ids": {
          "application_id": "app1"
        },
        "dev_eui": "4200000000000000",
        "join_eui": "4200000000000000"
      }
    }
  ],
  "data": {
    "@type": "type.googleapis.com/google.protobuf.Value",
    "value": 1 // There was 1 gateway that received the join-request.
  },
  "correlation_ids": [
    // Here you find the correlation IDs of all gs.up.receive events that were merged.
    "gs:conn:01D7KWADW2E5CJA32VS1MTR2J6",
    "gs:uplink:01D7KWB0N2KVCV8HZABC8DDHSA",
    "ns:uplink:01D7KWB0N5C1T8TE2HAVBJN5Y4",
    "rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D7KWB0N5G2N5C0AFXT4YMF8R"
  ]
}
{
  "name": "as.up.join.receive", // Application Server receives the join-accept.
  "time": "2019-04-04T09:54:35.005090Z",
  "identifiers": [
    {
      "device_ids": {
        "device_id": "dev1",
        "application_ids": {
          "application_id": "app1"
        },
        "dev_eui": "4200000000000000",
        "join_eui": "4200000000000000",
        "dev_addr": "0063ECE2"
      }
    }
  ],
  "correlation_ids": [
    "as:up:01D7KWB0VX1D7G3RKFN9HDA39Q",
    "gs:conn:01D7KWADW2E5CJA32VS1MTR2J6",
    "gs:uplink:01D7KWB0N2KVCV8HZABC8DDHSA",
    "ns:uplink:01D7KWB0N5C1T8TE2HAVBJN5Y4",
    "rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D7KWB0N5G2N5C0AFXT4YMF8R"
  ]
}
{
  "name": "as.up.join.forward", // Application Server forwards the join-accept to an application (CLI, MQTT, webhooks, etc).
  "time": "2019-04-04T09:54:35.010243Z",
  "identifiers": [
    {
      "device_ids": {
        "device_id": "dev1",
        "application_ids": {
          "application_id": "app1"
        },
        "dev_eui": "4200000000000000",
        "join_eui": "4200000000000000",
        "dev_addr": "0063ECE2"
      }
    }
  ],
  "correlation_ids": [
    "as:up:01D7KWB0VX1D7G3RKFN9HDA39Q",
    "gs:conn:01D7KWADW2E5CJA32VS1MTR2J6",
    "gs:uplink:01D7KWB0N2KVCV8HZABC8DDHSA",
    "ns:uplink:01D7KWB0N5C1T8TE2HAVBJN5Y4",
    "rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D7KWB0N5G2N5C0AFXT4YMF8R"
  ]
}
{
  "name": "gs.down.send", // Gateway Server sent the join-accept to the gateway.
  "time": "2019-04-04T09:54:35.046147Z",
  "identifiers": [
    {
      "gateway_ids": {
        "gateway_id": "multitech",
        "eui": "00800000A0000DB4"
      }
    }
  ],
  "correlation_ids": [
    "gs:conn:01D7KWADW2E5CJA32VS1MTR2J6",
    "rpc:/ttn.lorawan.v3.NsGs/ScheduleDownlink:01D7KWB0W84AJ1P5A3AQV6R4J7"
  ]
}
{
  "name": "gs.up.forward", // Gateway Server forwarded join-request to the Network Server.
  "time": "2019-04-04T09:54:35.991226Z",
  "identifiers": [
    {
      "gateway_ids": {
        "gateway_id": "multitech",
        "eui": "00800000A0000DB4"
      }
    }
  ],
  "correlation_ids": [
    "gs:conn:01D7KWADW2E5CJA32VS1MTR2J6",
    "gs:uplink:01D7KWB0N2KVCV8HZABC8DDHSA"
  ]
}
```
