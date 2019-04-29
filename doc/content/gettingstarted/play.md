---
title: "Play with things"
description: "Connect to the MQTT broker to receive uplinks and send downlinks"
weight: 6
--- 

## <a name="linkappserver">Linking the application</a>

In order to send uplinks and receive downlinks from your device, you must link the Application Server to the Network Server. In order to do this, create an API key for the Application Server:

```bash
$ ttn-lw-cli applications api-keys create \
  --name link \
  --application-id app1 \
  --right-application-link
```

The CLI will return an API key such as `NNSXS.VEEBURF3KR77ZR...`. This API key has only link rights and can therefore only be used for linking.

You can now link the Application Server to the Network Server:

```bash
$ ttn-lw-cli applications link set app1 --api-key NNSXS.VEEBURF3KR77ZR..
```

Your application is now linked. You can now use the builtin MQTT server and webhooks to receive uplink traffic and send downlink traffic.

## <a name="mqtt">Using the MQTT server</a>

In order to use the MQTT server you need to create a new API key to authenticate:

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
>`$ mosquitto_sub -h localhost -t '#' -u app1 -P 'NNSXS.VEEBURF3KR77ZR..' -d`

### Subscribing to messages

The Application Server publishes on the following topics:

- `v3/{application id}/devices/{device id}/join`
- `v3/{application id}/devices/{device id}/up`
- `v3/{application id}/devices/{device id}/down/queued`
- `v3/{application id}/devices/{device id}/down/sent`
- `v3/{application id}/devices/{device id}/down/ack`
- `v3/{application id}/devices/{device id}/down/nack`
- `v3/{application id}/devices/{device id}/down/failed`

While you could subscribe to separate topics, for the tutorial subscribe to `#` to subscribe to all messages.

With your MQTT client subscribed, when a device joins the network, a `join` message gets published. For example, for a device ID `dev1`, the message will be published on the topic `v3/app1/devices/dev1/join` with the following contents:

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

You can use the correlation IDs to follow messages as they pass through the stack.

When the device sends an uplink message, a message will be published to the topic `v3/{application id}/devices/{device id}/up`. This message looks like this:

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

### Scheduling a downlink message

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

#### Class C unicast

In order to send class C downlink messages to a single device, enable class C support for the end device using the following command:

```bash
$ ttn-lw-cli end-devices update app1 dev1 --supports-class-c
```

This will enable the class C downlink scheduling of the device. That's it! New downlink messages are now scheduled as soon as possible.

To disable class C scheduling, set reset with `--supports-class-c=false`.

>Note: you can also pass `--supports-class-c` when creating the device. Class C scheduling will be enable after the first uplink message which confirms the device session.

#### Class C multicast

Multicast messages are downlinks messages which are sent to multiple devices that share the same security context. In the Network Server, this is an ABP session. See [creating a device](#createdev) for learning how to create an ABP device.

Multicast sessions do not allow uplink. Therefore, you need to explicitly specify the gateway(s) to send messages from, using the `class_b_c` field:

```json
{
  "downlinks": [{
    "f_port": 15,
    "frm_payload": "vu8=",
    "priority": "NORMAL",
    "class_b_c": {
      "gateways": [{
        "gateway_ids": {
          "gateway_id": "gtw1"
        }
      }]
    }
  }]
}
```

>Note: if you specify multiple gateways, the Network Server will try the gateways in the order specified. The first gateway with no conflicts and no duty-cycle limitation will send the message.

### Listing the downlink queue

The stack keeps a queue of downlink messages. Applications can keep pushing downlink messages or replace the queue with a list of downlink messages.

You can see what is in the queue;

```bash
$ ttn-lw-cli end-devices downlink list app1 dev1
```
