# The Things Network Stack for LoRaWAN

## Introduction

This document is a guide for setting up The Things Network's Network Stack V3 in a private environment. If you already have some knowledge about how the backend works and if you are comfortable with a command line, this is the perfect place to start.

In this guide we will get everything up and running on your local machine (on `localhost`) using Docker.
 
## Table of Contents

1. [Dependencies](#dependencies)
2. [Configuration](#configuration)
3. [Running the stack](#running)
4. [Login using the CLI](#login)
5. [Registering a gateway](#registergtw)
6. [Registering an application](#registerapp)
7. [Registering a device](#registerdev)
8. [Linking the application](#linkappserver)
9. [Using the MQTT broker](#mqtt)
10. [Using WebHooks](#webhooks)

## <a name="dependencies">Dependencies</a>

### CLI and Stack

Get the latest packages or binaries for your operating system from the [releases](https://github.com/TheThingsNetwork/lorawan-stack/releases) page on GitHub.

### Certificates

By default, the Stack requires a `cert.pem` and `key.pem`, in order to to serve content over TLS.

+ To generate self-signed certificates, you can use the following command. This requires a [Go environment setup](../DEVELOPMENT.md#development-environment).

```bash
go run $(go env GOROOT)/src/crypto/tls/generate_cert.go -ca -host localhost && chmod 0444 ./key.pem
```

Keep in mind that self-signed certificates are not trusted by browsers and operating systems, and as such they will return a warning regarding this matter. In order to avoid such issues, we recommend [Let's Encrypt](https://letsencrypt.org/getting-started/).

## <a name="configuration">Configuration</a>

The Stack can be started without passing any [configuration](config.md).

You can refer to our [networking documentation](networking.md) for the default endpoints of the Stack.

### Frequency plans

By default, frequency plans are fetched by the stack from the [`TheThingsNetwork/lorawan-frequency-plans` repository](https://github.com/TheThingsNetwork/lorawan-frequency-plans). To set a new source:

+ `TTN_LW_FREQUENCY_PLANS_URL` allows you to serve frequency plans fetched from a HTTP server.

+ `TTN_LW_FREQUENCY_PLANS_DIRECTORY` allows you to serve frequency plans from a local directory.

## <a name="running">Running the stack</a>

You can run it using Docker, or container orchestration solutions. An example [Docker Compose configuration](../docker-compose.yml) is available in the repository:

```bash
$ docker-compose pull
$ docker-compose run --rm stack is-db init
$ docker-compose run --rm stack is-db create-admin-user --id admin --email admin@localhost
$ docker-compose run --rm stack is-db create-oauth-client --id cli --name "Command Line Interface" --owner admin --no-secret --redirect-uri 'http://localhost:11885/oauth/callback' --redirect-uri '/oauth/code'
$ docker-compose up
```

This will create an admin user `admin`, and also create the OAuth client used by the CLI.

## <a name="login">Login using the CLI</a>

The CLI needs to be logged on in order to create gateways, devices or API keys. You can use the following commands in a separate console session to login:

```bash
$ ttn-lw-cli login
```

A link will be provided to the OAuth login page where you can login using the credentials from the step ahead.

## <a name="registergtw">Registering a gateway</a>

By default, the stack allows unregistered gateways to connect, but without providing a default band. As such, it is highly recommended that each gateway is registered:

```bash
$ ttn-lw-cli gateway create gtw1 --user-id admin --frequency_plan_id EU_863_870 --gateway-eui 00800000A00009EF --enforce-duty-cycle
```

This creates a gateway `gtw1` with the frequency plan `EU_863_870` and EUI `00800000A00009EF` that respects duty-cycle limitations. For more options, you can use the `--help` flag. You can now connect your gateway to the stack.

## <a name="registerapp">Registering an application</a>

In order to register a device, the controlling application of the said device must be registered first:

```bash
$ ttn-lw-cli app create app1 --user-id admin
```

This creates an application `app1` for the user `admin`.

## <a name="registerdev">Registering a device</a>

You can now register an OTAA activated device to be used with the stack as follows:

```bash
$ ttn-lw-cli end-devices create app1 dev1 --dev-eui 0004A30B001C0530 --join-eui 800000000000000C --frequency_plan_id EU_863_870 --root_keys.app_key.key 752BAEC23EAE7964AF27C325F4C23C9A --lorawan_phy_version 1.0.2-b --lorawan_version 1.0.2
```

This will create an LoRa 1.0.2 end-device `dev1` with DevEUI `0004A30B001C0530`, AppEUI `800000000000000C` and AppKey
 `752BAEC23EAE7964AF27C325F4C23C9A`. After flashing the AppEUI and AppKey (which you can choose on your own), you should be able to join the private network.
If you wish to enable class C support for this device, you can add the `--supports-class-c` flag in the above command.

## <a name="linkappserver">Linking the application</a>

In order to send uplinks and receive downlinks from your device, you must first link the application server to the network server of the private network. In order to achieve this, first create an API key:

```bash
$ ttn-lw-cli app api-keys create --application-id app1 --right-application-link
```

The CLI will return an API key such as `NNSXS.VEEBURF3KR77ZRUF5JGCFIJQ4FLH5ELQXGR2SQQ.EKMEIDASX5EZTGOPDCZXGAXEHMD4FD2NAYTJERPD55VV3WAXADZQ`.
This API key has only linking rights, and should be used only during the linking process. 

You can now link the application server to the network server:

```bash
$ ttn-lw-cli app link set app1 --api-key [...]
```

Your application is now linked, and can use the built-in MQTT broker and Webhooks support.

## <a name="mqtt">Using the MQTT broker</a>

In order to use the MQTT broker it is necessary to register a new API key that will be used during the authentication process:

```bash
$ ttn-lw-cli app api-keys create --application-id app1 --right-application-traffic-down-write --right-application-traffic-read
```

Note that this new API key has full rights and can both receive uplinks and schedule downlinks. You can now login using an MQTT client using the username `app1` (the application name) and the newly generated API key as password.

### Subscribing to messages

MQTT topics provided by the built-in broker follow the format `v3/{application id}/devices/{device id}/{traffic type}`. While you could indeed subscribe for separate topics, for the purpose of this tutorial we will use the wilcard topic `#`, which provides all of the available messages of the application.

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

### Scheduling a downlink

Downlinks can be scheduled by publishing the message to the topic `v3/{application id}/devices/{device id}/down/push`. For example, if we want to send an unconfirmed downlink to the device `dev-simulator` with a payload of `BE EF` on port 15, we can use the topic `v3/app1/devices/dev-simulator/down/push` with the following contents:

```json
{
	"downlinks": [{
		"f_port": 15,
		"frm_payload": "vu8="
	}]
}
```

The payload is Base64 formatted, and it is possible to send multiple downlinks on a single push (since `downlinks` is an array).

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

## <a name="webhooks">Using WebHooks</a>

The WebHooks feature allows the application server to send application related messages to specific HTTP(S) endpoints. Creating a WebHook requires you to have an endpoint available as a message sink.

```bash
$ ttn-lw-cli app webhook set --application-id app1 --webhook-id wh1 --base-url https://example.com/lorahooks --join-accept.path "join" --format "json"
```

This will create an WebHook `wh1` for the application `app1` with a base URL `https://example.com/lorahooks` and a join path `join`. When a device of the application `app1` joins the network, the application server will do a `POST` request on the endpoint `https://example.com/lorahooks/join` with the following body:

```json
{
	"end_device_ids": {
		"device_id": "dev-simulator",
		"application_ids": {
			"application_id": "app1"
		},
		"dev_eui": "4200000000000000",
		"join_eui": "4200000000000000",
		"dev_addr": "01E9EF6A"
	},
	"correlation_ids": ["gs:conn:01D2CSNX7FJVKQPCVG612QF1TX", "gs:uplink:01D2CWCK40JJFVY0J9KXQ2QQYP", "ns:uplink:01D2CWCK41YNDZ16QX3MFY7YAT", "rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D2CWCK418QERZBCP7AXDGX4J", "as:up:01D2CWCKAB5B148AHB0ED5MCQE"],
	"join_accept": {
		"session_key_id": "AWiZxkyDbxhYoP22ceb7SQ=="
	}
}
```

You can later on subscribe for other messages, such as uplinks, using the following command:

```bash
$ ttn-lw-cli app webhook set --application-id app1 --webhook-id wh1 --uplink-message.path "up"
```

Now when the device sends an uplink, the application server will do a `POST` request to `https://example.com/lorahooks/up` with the following body:

```json
{
	"end_device_ids": {
		"device_id": "dev-simulator",
		"application_ids": {
			"application_id": "app1"
		},
		"dev_eui": "4200000000000000",
		"join_eui": "4200000000000000",
		"dev_addr": "01E9EF6A"
	},
	"correlation_ids": ["gs:conn:01D2CSNX7FJVKQPCVG612QF1TX", "gs:uplink:01D2CWX2MQBNBTE6M8TJ81K96K", "ns:uplink:01D2CWX2MQGRMBDEN88KHV3S5Z", "rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D2CWX2MQFAV25GK9M6WB5DM7", "as:up:01D2CWX2V1NBSBMDQYFFH4B2VR"],
	"uplink_message": {
		"session_key_id": "AWiZxkyDbxhYoP22ceb7SQ==",
		"f_port": 15,
		"frm_payload": "VGVtcGVyYXR1cmUgPSAwLjA=",
		"rx_metadata": [{
			"gateway_ids": {
				"gateway_id": "eui-0242020000247803",
				"eui": "0242020000247803"
			},
			"time": "2019-01-29T13:31:16.500Z",
			"timestamp": 3004844000,
			"rssi": -35,
			"snr": 5,
			"uplink_token": "CiIKIAoUZXVpLTAyNDIwMjAwMDAyNDc4MDMSCAJCAgAAJHgDEOCP6ZgL"
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
			"frequency": "868100000",
			"gateway_channel_index": 2
		}
	}
}
``` 
