---
title: "Scheduling a downlink message"
description: ""
weight: 1
---

Downlinks can be scheduled by publishing the message to the topic `v3/{application id}/devices/{device id}/down/push`.

For example, to send an unconfirmed downlink message to the device `dev1` in application `app1` with the hexadecimal payload `BE EF` on `FPort` 15 with normal priority, use the topic `v3/app1/devices/dev1/down/push` with the following contents:

```json
{
  "downlinks": [{
    "f_port": 15,
    "frm_payload": "vu8=",
    "priority": "NORMAL",
  }]
}
```

>Hint: Use [this handy tool](https://v2.cryptii.com/hexadecimal/base64) to convert hexadecimal to base64.

>If you use `mosquitto_pub`, use the following command:
>
>`$ mosquitto_pub -h localhost -t 'v3/app1/devices/dev1/down/push' -u app1 -P 'NNSXS.VEEBURF3KR77ZR..' -m '{"downlinks":[{"f_port": 15,"frm_payload":"vu8=","priority": "NORMAL"}]}' -d`

It is also possible to send multiple downlink messages on a single push because `downlinks` is an array. Instead of `/push`, you can also use `/replace` to replace the downlink queue. Replacing with an empty array clears the downlink queue.

>Note: if you do not specify a priority, the default priority `LOWEST` is used. You can specify `LOWEST`, `LOW`, `BELOW_NORMAL`, `NORMAL`, `ABOVE_NORMAL`, `HIGH` and `HIGHEST`.

The stack supports some cool features, such as confirmed downlink with your own correlation IDs. For example, you can push this:

```json
{
  "downlinks": [{
    "f_port": 15,
    "frm_payload": "vu8=",
    "priority": "HIGH",
    "confirmed": true,
    "correlation_ids": ["my-correlation-id"]
  }]
}
```

Once the downlink gets acknowledged, a message is published to the topic `v3/{application id}/devices/{device id}/down/ack`:

```json
{
  "end_device_ids": {
    "device_id": "dev1",
    "application_ids": {
      "application_id": "app1"
    },
    "dev_eui": "4200000000000000",
    "join_eui": "4200000000000000",
    "dev_addr": "00E6F42A"
  },
  "correlation_ids": [
    "my-correlation-id",
    "..."
  ],
  "downlink_ack": {
    "session_key_id": "AWnj0318qrtJ7kbudd8Vmw==",
    "f_port": 15,
    "f_cnt": 11,
    "frm_payload": "vu8=",
    "confirmed": true,
    "priority": "NORMAL",
    "correlation_ids": [
      "my-correlation-id",
      "..."
    ]
  }
}
```

Here you see the correlation ID `my-correlation-id` of your downlink message. You can add multiple custom correlation IDs, for example to reference events or identifiers of your application.
