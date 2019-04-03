# Getting Started with The Things Network Stack for LoRaWAN

## Introduction

This is a guide for setting up a private LoRaWAN Network Server using The Things Network Stack for LoRaWAN.

In this guide we will get everything up and running on a server using Docker. If you are comfortable with configuring servers and working with command line, this is the perfect place to start.
 
## Table of Contents

1. [Dependencies](#dependencies)
2. [Configuration](#configuration)
3. [Running the stack](#running)
4. [Login using the CLI](#login)
5. [Creating a gateway](#creategtw)
6. [Creating an application](#createapp)
7. [Creating a device](#createdev)
8. [Linking the application](#linkappserver)
9. [Using the MQTT server](#mqtt)
10. [Using webhooks](#webhooks)

## <a name="dependencies">Dependencies</a>

### CLI and stack

The web interface Console is not yet available. So in this tutorial, we use the command-line interface (CLI) to manage the setup.

You can use the CLI on your local machine or on the server.

#### Package managers (recommended)

##### macOS

```bash
$ brew install TheThingsNetwork/lorawan-stack/ttn-lw-stack
```

##### Ubuntu

```bash
$ sudo snap install ttn-lw-stack
```

#### Binaries

If your operating system or package manager is not mentioned, please [download binaries](https://github.com/TheThingsNetwork/lorawan-stack/releases) for your operating system and processor architecture.

### Certificates

By default, the stack requires a `cert.pem` and `key.pem`, in order to to serve content over TLS.

+ To generate self-signed certificates for `localhost`, use the following commands. This requires a [Go environment setup](../DEVELOPMENT.md#development-environment).

```bash
$ go run $(go env GOROOT)/src/crypto/tls/generate_cert.go -ca -host localhost 
# The following command is not required on Windows.
$ chmod 0444 ./key.pem
```

Keep in mind that self-signed certificates are not trusted by browsers and operating systems, resulting in warnings and sometimes in errors. Consider [Let's Encrypt](https://letsencrypt.org/getting-started/) for free and trusted TLS certificates for your server.

## <a name="configuration">Configuration</a>

The stack can be started without passing any configuration. However, there are a lot of things you can configure. See [configuration documentation](config.md) for more information.

Refer to the [networking documentation](networking.md) for the endpoints and ports that the stack uses by default.

### <a name="frequencyplans">Frequency plans</a>

By default, frequency plans are fetched by the stack from a [public GitHub repository](https://github.com/TheThingsNetwork/lorawan-frequency-plans). To configure a local directory in offline environments, see the [configuration documentation](config.md) for more information.

## <a name="running">Running the stack</a>

You can run the stack using Docker or container orchestration solutions using Docker. An example [Docker Compose configuration](../docker-compose.yml) is available to get started quickly.

With the `docker-compose.yml` file in the directory of your terminal prompt, enter the following commands to initialize the database, create the first user `admin` with password `admin`, create the CLI OAuth client and start the stack:

```bash
$ docker-compose pull
$ docker-compose run --rm stack is-db init
$ docker-compose run --rm stack is-db create-admin-user
  --id admin \
  --password admin \
  --email admin@localhost
$ docker-compose run --rm stack is-db create-oauth-client \
  --id cli \
  --name "Command Line Interface" \
  --owner admin \
  --no-secret \
  --redirect-uri 'local-callback' \
  --redirect-uri 'code'
$ docker-compose up
```

## <a name="login">Login using the CLI</a>

The CLI needs to be logged on in order to create gateways, applications, devices and API keys. With the stack running in one terminal session, login with the following command:

```bash
$ ttn-lw-cli login
```

A link will be provided to the OAuth login page where you can login using the credentials from the step ahead. Once you logged in in the browser, return to the terminal session to proceed.

## <a name="creategtw">Creating a gateway</a>

Create the first gateway:

```bash
$ ttn-lw-cli gateway create gtw1 \
  --user-id admin \
  --frequency-plan-id EU_863_870 \
  --gateway-eui 00800000A00009EF \
  --enforce-duty-cycle
```

This creates a gateway `gtw1` with the frequency plan `EU_863_870`, EUI `00800000A00009EF`, respecting duty-cycle limitations and with the `admin` user as collaborator. You can now connect your gateway to the stack.

>Note: if you need help with any command of `ttn-lw-cli`, use the `--help` flag to get a list of subcommands, flags and their description and aliases.

## <a name="createapp">Creating an application</a>

Create the first application:

```bash
$ ttn-lw-cli app create app1 --user-id admin
```

This creates an application `app1` with the `admin` user as collaborator.

Devices are created within applications.

## <a name="createdev">Creating a device</a>

Creating a device with over-the-air-activation (OTAA) to be used with the stack:

```bash
$ ttn-lw-cli end-devices create app1 dev1 \
  --dev-eui 0004A30B001C0530 \
  --app-eui 800000000000000C \
  --frequency-plan-id EU_863_870 \
  --root-keys.app-key.key 752BAEC23EAE7964AF27C325F4C23C9A \
  --lorawan-version 1.0.2 \
  --lorawan-phy-version 1.0.2-b
```

This will create a LoRaWAN 1.0.2 end device `dev1` in application `app1`. The end device should now be able to join the private network.

It is also possible to register an ABP activated device using the `--abp` flag as follows:

```bash
$ ttn-lw-cli end-devices create app1 dev2 --frequency-plan-id EU_863_870 --lorawan-version 1.0.2 --lorawan-phy-version 1.0.2-b --abp --session.dev-addr 00E4304D --session.keys.app-s-key.key A0CAD5A30036DBE03096EB67CA975BAA --session.keys.nwk_s_key.key B7F3E161BC9D4388E6C788A0C547F255
```

## <a name="linkappserver">Linking the application</a>

In order to send uplinks and receive downlinks from your device, you must link the Application Server to the Network Server. In order to do this, create an API key for the Application Server:

```bash
$ ttn-lw-cli app api-keys create \
  --application-id app1 \
  --right-application-link
```

The CLI will return an API key such as `NNSXS.VEEBURF3KR77ZR...`. This API key has only link rights and can therefore only be used for linking.

You can now link the Application Server to the Network Server:

```bash
$ ttn-lw-cli app link set app1 --api-key NNSXS.VEEBURF3KR77ZR..
```

Your application is now linked. You can now use the builtin MQTT server and webhooks to receive uplink traffic and send downlink traffic.

## <a name="mqtt">Using the MQTT server</a>

In order to use the MQTT server you need to create a new API key to authenticate:

```bash
$ ttn-lw-cli app api-keys create \
  --application-id app1 \
  --right-application-traffic-read \
  --right-application-traffic-down-write
```

You can now login using an MQTT client with the application ID `app1` as user name and the newly generated API key as password.

There are many MQTT clients available. Great clients are `mosquitto_pub` and `mosquitto_sub`, part of [Mosquitto](https://mosquitto.org).

>Tip: when using `mosquitto_sub`, pass the `-d` flag to see the topics messages get published on. For example:
>
>`$ mosquitto_sub -h localhost -t '#' -u app1 -P 'NNSXS.VEEBURF3KR77ZR..' -d`

### Subscribing to messages

MQTT topics provided by the Application Server follow the format `v3/{application id}/devices/{device id}/{message type}`. While you could subscribe to separate topics, in this tutorial we use the wildcard topic `#` to subscribe to all messages.

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

>If you use `mosquitto_pub`, use the following command:
>
>`$ mosquitto_pub -h localhost -t 'v3/app1/devices/dev1/down/push' -u app1 -P 'NNSXS.VEEBURF3KR77ZR..' -m '{"downlinks":[{"f_port": 15,"frm_payload":"vu8=","priority": "NORMAL",}]}' -d`

The payload is base64 formatted, and it is possible to send multiple downlinks on a single push (since `downlinks` is an array). Instead of `push`, you can also use `replace` to replace the downlink queue.

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
    "as:conn:01D7HYP6YKJTZ9ADK25ESR2D33",
    "as:downlink:01D7HYP6YMR93065JQ2M5SCQ95",
    "gs:conn:01D7HYP23MMAW0Z5DDX8H1GS2K",
    "gs:uplink:01D7HYP8CAG29MNE5VAY4P39WS",
    "ns:uplink:01D7HYP8CDZNDQY6TA9R5PJHDW",
    "rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01D7HYP8CDYT09PJEW9BKP51BG",
    "as:up:01D7HYP8CF8CN5ZK2D2MQB7MD1"
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
      "as:conn:01D7HYP6YKJTZ9ADK25ESR2D33",
      "as:downlink:01D7HYP6YMR93065JQ2M5SCQ95"
    ]
  }
}
```

Here you see the correlation ID `my-correlation-id` of your downlink message. You can add multiple custom correlation IDs, for example to reference events or identifiers of your application.

#### Class C unicast

In order to send class C downlink messages to a single device, enable class C support for the end device using the following command:

```bash
$ ttn-lw-cli end-devices set app1 dev1 --supports-class-c
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

## Listing the downlink queue

The stack keeps a queue of downlink messages. Applications can keep pushing downlink messages or replace the queue with a list of downlink messages.

You can see what is in the queue;

```bash
$ ttn-lw-cli end-devices downlink list app1 dev1
```

## <a name="webhooks">Using webhooks</a>

The webhooks feature allows the Application Server to send application related messages to specific HTTP(S) endpoints. The `json` formatter uses the same format as the MQTT server described above.

Creating a webhook requires you to have an endpoint available as a message sink.

```bash
$ ttn-lw-cli applications webhook set --application-id app1 \
  --webhook-id wh1 \
  --format json \
  --base-url https://example.com/lorahooks \
  --join-accept.path /join \
  --uplink-message.path /up
```

This will create an webhook `wh1` for the application `app1` with a base URL `https://example.com/lorahooks` with JSON formatting. The Application Server performs `POST` requests on the endpoint `https://example.com/lorahooks/join` for join-accepts and `https://example.com/lorahooks/up` for uplink messages.

>Note: You can also specify URL paths for downlink events, see `ttn-lw-cli app webhook set --help` for more information.

You can also send downlink messages through using webhooks. The path is `/v3/api/as/applications/{application_id}/webhooks/{webhook_id}/devices/{device_id}/down/push` (or `/replace`). Pass the API key as
bearer token on the `Authorization` header. For example:

```
$ curl http://localhost:1885/v3/api/as/applications/app1/webhooks/wh1/devices/dev1/down/push \
  -X POST \
  -H 'Authorization: Bearer NNSXS.VEEBURF3KR77ZR..' \
  --data '{"downlinks":[{"frm_payload":"vu8=","f_port":15,"priority":"NORMAL"}]}'
```

## Congratulations

You have now set up The Things Network Stack V3! 🎉

## Advanced: Events

The stack generates lots of events that allow you to get insight in what is going on. You can subscribe to application, gateway, end device events, as well as to user, organization and OAuth client events.

### Using the CLI

To follow your gateway `gtw1` and application `app1` events at the same time:

```bash
$ ttn-lw-cli events subscribe --gateway-id gtw1 --application-id app1
```

### Using cURL

You can also get streaming events with `curl`. For this, you need an API key for the entities you want to watch, for example:

```bash
$ ttn-lw-cli user api-key create --user-id admin --right-application-all --right-gateway-all
```

With the created API key:

```
$ curl http://localhost:1885/api/v3/events \
  -X POST
  -H 'Authorization: Bearer NNSXS.BR55PTYILPPVXY..' \
  --data '{"identifiers":[{"application_ids":{"application_id":"app1"}},{"gateway_ids":{"gateway_id":"gtw1"}}]}'
```

### Example: join flow

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
