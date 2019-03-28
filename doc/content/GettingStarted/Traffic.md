---
title: "Device interaction"
description: "Connect to the MQTT broker to receive uplinks and send downlinks"
weight: 4
draft: false
--- 

## <a name="mqtt">Using the MQTT broker</a>

In order to use the MQTT broker it is necessary to register a new API key that will be used during the authentication process:

```bash
$ ttn-lw-cli app api-keys create --application-id app1 --right-application-traffic-down-write --right-application-traffic-read
```

Note that this new API key can both receive uplinks and schedule downlinks.
You can now login using an MQTT client using the username `app1` (the application name) and the newly generated API key as password.

There are many MQTT clients available; a simple one is `mosquitto_pub` and `mosquitto_sub`, part of [Mosquitto](https://mosquitto.org).

### Subscribing to messages

MQTT topics provided by the built-in broker follow the format `v3/{application id}/devices/{device id}/{traffic type}`. While you could indeed subscribe for separate topics, for the purpose of this tutorial we will use the wildcard topic `#`, which provides all of the available messages of the application.

After subscribing to `#` from your client, when a device of the application that is currently logged in joins the network, a `join` message will be published. For example, for a device called `dev-simulator`, the message will be published on the topic `v3/app1/devices/dev-simulator/join` with the following contents:

```json
{
	"end_device_ids": {
		"device_id": "dev-simulator",
		"application_ids": {
			"application_id": "app1"
		},
		"dev_eui": "4200000000000000",
		"join_eui": "4200000000000000",
		"dev_addr": "01DA1F15"
	},
	"correlation_ids": ["gs:conn:01D2CSNX7FJVKQPCVG612QF1TX", "gs:uplink:01D2CT834K2YD17ZWZ6357HC0Z", "ns:uplink:01D2CT834KNYD7BT2NHK5R1WVA", "rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D2CT834KJ4AVSD1SJ637NAV6", "as:up:01D2CT83AXQFQYQ35SR74CTWKH"],
	"join_accept": {
		"session_key_id": "AWiZpAyXrAfEkUNkBljRoA=="
	}
}
```

As you can see, with correlation IDs it will be possible to follow each message as it passes through the stack components, which can be handy while debugging.

When the device sends an uplink, the message will be broadcasted to the topic `v3/app1/devices/dev-simulator/up` and will contain a payload formatted as follows:

```json
{
	"end_device_ids": {
		"device_id": "dev-simulator",
		"application_ids": {
			"application_id": "app1"
		},
		"dev_eui": "4200000000000000",
		"join_eui": "4200000000000000",
		"dev_addr": "01DA1F15"
	},
	"correlation_ids": ["gs:conn:01D2CSNX7FJVKQPCVG612QF1TX", "gs:uplink:01D2CV8HF62ME0D7MZWE38HHH8", "ns:uplink:01D2CV8HF6FYJHKZ45YY1DB3MR", "rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D2CV8HF6XR7ZFVK768PDG3J4", "as:up:01D2CV8HNGJ57G25BW0FCZNY07"],
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

#### Class A downlinks

Downlinks can be scheduled by publishing the message to the topic `v3/{application id}/devices/{device id}/down/push`. For example, if we want to send an unconfirmed downlink to the device `dev-simulator` with a payload of `BE EF` on port 15, we can use the topic `v3/app1/devices/dev-simulator/down/push` with the following contents:

```json
{
	"downlinks": [{
		"f_port": 15,
		"frm_payload": "vu8="
	}]
}
```

The payload is base64 formatted, and it is possible to send multiple downlinks on a single push (since `downlinks` is an array). Instead of `push`, you can also use `replace` to replace the downlink queue.

If we want to send a confirmed downlink to our device, we will use the same topic but add the `confirmed` flag to the downlink.

```json
{
	"downlinks": [{
		"f_port": 15,
		"frm_payload": "vu8=",
		"confirmed": true
	}]
}
```

Once the downlink has been acknowledged, a message is published to the topic `v3/app1/devices/dev-simulator/down/ack`:

```json
{
	"end_device_ids": {
		"device_id": "dev-simulator",
		"application_ids": {
			"application_id": "app1"
		},
		"dev_eui": "4200000000000000",
		"join_eui": "4200000000000000",
		"dev_addr": "01DA1F15"
	},
	"correlation_ids": ["as:conn:01D2CT5BZNX862RP9SV2JSRWZ7", "as:downlink:01D2CVN6WW9S152ZVB0C7VHM4Z", "gs:conn:01D2CSNX7FJVKQPCVG612QF1TX", "gs:uplink:01D2CVQP4ZW7BSFHFCCP8ECHY4", "ns:uplink:01D2CVQP502TQ417XPWDZHKH41", "rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D2CVQP50QD90AJCH80N4KKYM", "as:up:01D2CVQP51P18F6VGXZMG0EXGS"],
	"downlink_ack": {
		"session_key_id": "AWiZpAyXrAfEkUNkBljRoA==",
		"f_port": 15,
		"f_cnt": 8,
		"frm_payload": "vu8=",
		"confirmed": true,
		"correlation_ids": ["as:conn:01D2CT5BZNX862RP9SV2JSRWZ7", "as:downlink:01D2CVN6WW9S152ZVB0C7VHM4Z"]
	}
}
```

#### Class C downlinks

In order to schedule class C downlinks, the support for class C scheduling has to be enabled in the network server using the following command:

```bash
$ ttn-lw-cli end-devices set app1 dev1 --supports-class-c
```

This will enable the class C downlink scheduling of the device. It is assumed that devices with LoRaWAN versions earlier than 1.1 enable class C after the join procedure, while later devices use the `DeviceMode` MAC command to change their own class.

No other changes are required in the format of the downlink message, since class C support is related to downlink scheduling.

#### Class C multicast downlinks

Multicast downlinks are downlinks which are sent to a specific ABP session which is shared by multiple devices. Since the session is shared by multiple devices, the downlink will be received by all of them when it's transmitted by the gateway. 

Class C scheduling support is required in order to achieve this, and can be enabled using the command in the section above.

Downlinks can be scheduled by adding the `class_bc` flag to the downlink message, which specifies on which gateway(s) the downlink should be scheduled.

```json
{
    "downlinks": [{
        "f_port": 15,
        "frm_payload": "vu8=",
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
