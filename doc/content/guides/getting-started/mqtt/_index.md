---
title: "MQTT Server"
description: ""
weight: 20
---

The Application Server exposes a MQTT server to work with streaming events. In order to use the MQTT server you need to create a new API key to authenticate:

```bash
$ ttn-lw-cli applications api-keys create \
  --name mqtt-client \
  --application-id app1 \
  --right-application-traffic-read \
  --right-application-traffic-down-write
```

>Note: See `--help` to see more rights that your application may need.

You can now login using an MQTT client with the application ID `app1` as user name and the newly generated API key as password.

There are many MQTT clients available. Great clients are `mosquitto_pub` and `mosquitto_sub`, part of [Mosquitto](https://mosquitto.org).

>Tip: when using `mosquitto_sub`, pass the `-d` flag to see the topics messages get published on. For example:
>
>```bash
$ mosquitto_sub -h localhost -t '#' -u app1 -P 'NNSXS.VEEBURF3KR77ZR..' -d
```

## Subscribing to upstream traffic

The Application Server publishes on the following topics:

- `v3/{application id}/devices/{device id}/join`
- `v3/{application id}/devices/{device id}/up`
- `v3/{application id}/devices/{device id}/down/queued`
- `v3/{application id}/devices/{device id}/down/sent`
- `v3/{application id}/devices/{device id}/down/ack`
- `v3/{application id}/devices/{device id}/down/nack`
- `v3/{application id}/devices/{device id}/down/failed`

While you could subscribe to separate topics, for the tutorial subscribe to `#` to subscribe to all messages.

With your MQTT client subscribed, when a device joins the network, a `join` message gets published. For example, for a device ID `dev1`, the message will be published on the topic `v3/app1/devices/dev1/join`.

<details><summary>Show example</summary>
```json
{
  "end_device_ids": {
    "device_id": "dev1",
    "application_ids": {
      "application_id": "app1"
    },
    "dev_eui": "4200000000000000",
    "join_eui": "4200000000000000",
    "dev_addr": "01DA1F15"
  },
  "correlation_ids": [
    "gs:conn:01D2CSNX7FJVKQPCVG612QF1TX",
    "gs:uplink:01D2CT834K2YD17ZWZ6357HC0Z",
    "ns:uplink:01D2CT834KNYD7BT2NHK5R1WVA",
    "rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D2CT834KJ4AVSD1SJ637NAV6",
    "as:up:01D2CT83AXQFQYQ35SR74CTWKH"
  ],
  "join_accept": {
    "session_key_id": "AWiZpAyXrAfEkUNkBljRoA=="
  }
}
```
</details>

You can use the correlation IDs to follow messages as they pass through The Things Stack.

When the device sends an uplink message, a message will be published to the topic `v3/{application id}/devices/{device id}/up`.

<details><summary>Show example</summary>
```json
{
  "end_device_ids": {
    "device_id": "dev1",
    "application_ids": {
      "application_id": "app1"
    },
    "dev_eui": "4200000000000000",
    "join_eui": "4200000000000000",
    "dev_addr": "01DA1F15"
  },
  "correlation_ids": [
    "gs:conn:01D2CSNX7FJVKQPCVG612QF1TX",
    "gs:uplink:01D2CV8HF62ME0D7MZWE38HHH8",
    "ns:uplink:01D2CV8HF6FYJHKZ45YY1DB3MR",
    "rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D2CV8HF6XR7ZFVK768PDG3J4",
    "as:up:01D2CV8HNGJ57G25BW0FCZNY07"
  ],
  "uplink_message": {
    "session_key_id": "AWiZpAyXrAfEkUNkBljRoA==",
    "f_port": 15,
    "frm_payload": "VGVtcGVyYXR1cmUgPSAwLjA=",
    "rx_metadata": [{
      "gateway_ids": {
        "gateway_id": "eui-0242020000247803",
        "eui": "0242020000247803"
      },
      "time": "2019-01-29T13:02:34.981Z",
      "timestamp": 1283325000,
      "rssi": -35,
      "snr": 5,
      "uplink_token": "CiIKIAoUZXVpLTAyNDIwMjAwMDAyNDc4MDMSCAJCAgAAJHgDEMj49+ME"
    }],
    "settings": {
      "data_rate": {
        "lora": {
          "bandwidth": 125000,
          "spreading_factor": 7
        }
      },
      "data_rate_index": 5,
      "coding_rate": "4/6",
      "frequency": "868500000",
      "gateway_channel_index": 2,
      "device_channel_index": 2
    }
  }
}
```
</details>

## Publishing downlink traffic

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
>```bash
$ mosquitto_pub -h localhost \
  -t 'v3/app1/devices/dev1/down/push' \
  -u app1 -P 'NNSXS.VEEBURF3KR77ZR..' \
  -m '{"downlinks":[{"f_port": 15,"frm_payload":"vu8=","priority": "NORMAL"}]}' \
  -d`
```

It is also possible to send multiple downlink messages on a single push because `downlinks` is an array. Instead of `/push`, you can also use `/replace` to replace the downlink queue. Replacing with an empty array clears the downlink queue.

>Note: if you do not specify a priority, the default priority `LOWEST` is used. You can specify `LOWEST`, `LOW`, `BELOW_NORMAL`, `NORMAL`, `ABOVE_NORMAL`, `HIGH` and `HIGHEST`.

The Things Stack supports some cool features, such as confirmed downlink with your own correlation IDs. For example, you can push this:

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

Once the downlink gets acknowledged, a message is published to the topic `v3/{application id}/devices/{device id}/down/ack`.

<details><summary>Show example</summary>
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
</details>

You see the correlation ID `my-correlation-id` of your downlink message. You can add multiple custom correlation IDs, for example to reference events or identifiers of your application.
