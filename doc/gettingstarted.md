# Getting Started with The Things Network Stack for LoRaWAN

## Introduction

This is a guide for setting up a private LoRaWAN network using The Things Network Stack for LoRaWAN V3 on a server. If you already have some knowledge about how the stack works and if you are comfortable with a command line, this is the perfect place to start.

In this guide we will get everything up and running on a server using Docker.
 
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
10. [Using webhooks](#webhooks)

## <a name="dependencies">Dependencies</a>

### CLI and Stack

For now, the web interface Console is not available; in this tutorial, we use the command-line interface (CLI) to manage the network.

In this tutorial, the stack will run in Docker on the server, and you manage the stack either using the CLI on your local machine or on the server.

#### Package managers (recommended)

##### macOS

```bash
$ brew install TheThingsNetwork/lorawan-stack/ttn-lw-stack
```

#### Binaries

If your operating system or package manager is not mentioned, please [download binaries](https://github.com/TheThingsNetwork/lorawan-stack/releases) for your operating system and processor architecture.

### Certificates

By default, the Stack requires a `cert.pem` and `key.pem`, in order to to serve content over TLS.

+ To generate self-signed certificates for `localhost`, use the following commands. This requires a [Go environment setup](../DEVELOPMENT.md#development-environment).

```bash
$ go run $(go env GOROOT)/src/crypto/tls/generate_cert.go -ca -host localhost 
# The following command is not required on Windows.
$ chmod 0444 ./key.pem
```

Keep in mind that self-signed certificates are not trusted by browsers and operating systems, resulting oftentimes in warnings and sometimes in errors. Consider [Let's Encrypt](https://letsencrypt.org/getting-started/) for free and trusted TLS certificates for your server.

## <a name="configuration">Configuration</a>

The Stack can be started without passing any [configuration](config.md).

You can refer to our [networking documentation](networking.md) for the default endpoints of the Stack.

### <a name="frequencyplans">Frequency plans</a>

By default, frequency plans are fetched by the stack from the [`TheThingsNetwork/lorawan-frequency-plans` repository](https://github.com/TheThingsNetwork/lorawan-frequency-plans). To set a new source:

+ `TTN_LW_FREQUENCY_PLANS_URL` allows you to serve frequency plans fetched from a HTTP server.

+ `TTN_LW_FREQUENCY_PLANS_DIRECTORY` allows you to serve frequency plans from a local directory.

## <a name="running">Running the stack</a>

You can run the stack using Docker or container orchestration solutions using Docker. An example [Docker Compose configuration](../docker-compose.yml) is available in the repository.

With the `docker-compose.yml` file in the directory of your terminal prompt, enter the following commands to initialize the database, create the first user `admin`, create the CLI OAuth client and start the stack:

```bash
$ docker-compose pull
$ docker-compose run --rm stack is-db init
$ docker-compose run --rm stack is-db create-admin-user --id admin --email admin@localhost
$ docker-compose run --rm stack is-db create-oauth-client --id cli --name "Command Line Interface" --owner admin --no-secret --redirect-uri 'http://localhost:11885/oauth/callback' --redirect-uri 'code'
$ docker-compose up
```

## <a name="login">Login using the CLI</a>

The CLI needs to be logged on in order to create gateways, applications, devices and API keys. With the stack running in one terminal session, login with the following command:

```bash
$ ttn-lw-cli login
```

A link will be provided to the OAuth login page where you can login using the credentials from the step ahead. Once you logged in in the browser, return to the terminal session to proceed.

## <a name="registergtw">Registering a gateway</a>

By default, the stack allows unregistered gateways to connect, but without providing a default band. As such, it is highly recommended that each gateway is registered:

```bash
$ ttn-lw-cli gateway create gtw1 --user-id admin --frequency-plan-id EU_863_870 --gateway-eui 00800000A00009EF --enforce-duty-cycle
```

This creates a gateway `gtw1` with the frequency plan `EU_863_870` and EUI `00800000A00009EF` that respects duty-cycle limitations. You can now connect your gateway to the stack.

The frequency plan is fetched automatically from the [configured source](#frequencyplans).

>Note: if you need help with any command in `ttn-lw-cli`, use the `--help` flag to get a list of subcommands, flags and their description and aliases.

## <a name="registerapp">Registering an application</a>

In order to register a device, an application must be created first:

```bash
$ ttn-lw-cli app create app1 --user-id admin
```

This creates an application `app1` for the user `admin`.

## <a name="registerdev">Registering a device</a>

You can now register an OTAA activated device to be used with the stack as follows:

```bash
$ ttn-lw-cli end-devices create app1 dev1 --dev-eui 0004A30B001C0530 --join-eui 800000000000000C --frequency-plan-id EU_863_870 --root-keys.app-key.key 752BAEC23EAE7964AF27C325F4C23C9A --lorawan-phy-version 1.0.2-b --lorawan-version 1.0.2
```

This will create an LoRaWAN 1.0.2 end device `dev1` with DevEUI `0004A30B001C0530`, AppEUI `800000000000000C` and AppKey `752BAEC23EAE7964AF27C325F4C23C9A`. After configuring the credentials in the end device, you should be able to join the private network.

It is also possible to register an ABP activated device using the `--abp` flag as follows:

```bash
$ ttn-lw-cli end-devices create app1 dev1 --frequency-plan-id EU_863_870 --lorawan-phy-version 1.0.2-b --lorawan-version 1.0.2 --abp --session.dev-addr 00E4304D --session.keys.app-s-key.key A0CAD5A30036DBE03096EB67CA975BAA --session.keys.f_nwk_s_int_key.key B7F3E161BC9D4388E6C788A0C547F255
```

This will create an LoRaWAN 1.0.2 end device `dev1` with DevAddr `00E4304D`, AppSKey `A0CAD5A30036DBE03096EB67CA975BAA` and NwkSKey `B7F3E161BC9D4388E6C788A0C547F255`.

## <a name="linkappserver">Linking the application</a>

In order to send uplinks and receive downlinks from your device, you must first link the application server to the network server. In order to do this, first create an API key for the application server:

```bash
$ ttn-lw-cli app api-keys create --application-id app1 --right-application-link
```

The CLI will return an API key such as `NNSXS.VEEBURF3KR77ZR...`. This API key has only link rights and can therefore only be used during the linking process. 

You can now link the application server to the network server:

```bash
$ ttn-lw-cli app link set app1 --api-key NNSXS.VEEBURF3KR77ZR..
```

Your application is now linked, and can use the built-in MQTT broker and webhooks support to receive uplink traffic and send downlink traffic.

## <a name="mqtt">Using the MQTT broker</a>

In order to use the MQTT broker it is necessary to register a new API key that will be used during the authentication process:

```bash
$ ttn-lw-cli app api-keys create --application-id app1 --right-application-traffic-down-write --right-application-traffic-read
```

Note that this new API key can both receive uplinks and schedule downlinks. You can now login using an MQTT client using the username `app1` (the application name) and the newly generated API key as password.

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

## <a name="webhooks">Using webhooks</a>

The webhooks feature allows the application server to send application related messages to specific HTTP(S) endpoints. Creating a webhook requires you to have an endpoint available as a message sink.

```bash
$ ttn-lw-cli app webhook set --application-id app1 --webhook-id wh1 --base-url https://example.com/lorahooks --join-accept.path "join" --format "json"
```

This will create an webhook `wh1` for the application `app1` with a base URL `https://example.com/lorahooks` and a join-accept path `join`. When a device of the application `app1` joins the network, the application server will do a `POST` request on the endpoint `https://example.com/lorahooks/join` with the following body:

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

## Congratulations

You have now set up The Things Network Stack V3! 🎉
