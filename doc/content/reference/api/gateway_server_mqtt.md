---
title: "Gateway Server MQTT"
description: ""
---

This reference describes the MQTT protocol used by the Gateway Server. Packet forwarders implementing the MQTT protocols are specific for {{% tts %}}.

The MQTT protocol can be used to develop custom packet forwarders or gateway bridges for exchanging traffic between a gateway and {{% tts %}}, or easily simulating gateway traffic for testing purposes. It is an alternative to the [Basic Station](https://doc.sm.tc/station/index.html) and the [Semtech UDP](https://github.com/Lora-net/packet_forwarder/blob/master/PROTOCOL.TXT) protocols.

## MQTT Introduction

MQTT is a server-client protocol for exchanging messages. Clients can connect to the server and **publish** messages (data) under a specific **topic**. They can also **subscribe** to a **topic**, and thus receive all messages that are published under that topic (by other clients, or the MQTT server itself).

The Gateway Server is the MQTT server, and gateways connect to the Gateway Server as MQTT clients. The gateway that connects is identified by the username that is used for authentication.

## Protocol Buffers

To communicate with the MQTT protocol, the Gateway Server and the gateway are exchanging [Protocol Buffers](https://developers.google.com/protocol-buffers). The definitions of the Protocol Buffers can be found at the [GitHub repository](https://github.com/TheThingsNetwork/lorawan-stack) of {{% tts %}}, under [**messages.proto**](https://github.com/TheThingsNetwork/lorawan-stack/blob/default/api/messages.proto) and [**lorawan.proto**](https://github.com/TheThingsNetwork/lorawan-stack/blob/default/api/lorawan.proto).

## Connecting to the Gateway Server

See [Networking]({{< ref "/reference/networking" >}}) for the default port of the MQTT server.

The username is `<gateway-id>`, and the password is a gateway API key with the `RIGHT_GATEWAY_LINK` right enabled. You can generate this API key using the [Console]({{< ref "/getting-started/console/create-gateway#create-gateway-api-key" >}}) or the [CLI]({{< ref "/getting-started/cli/create-gateway#create-gateway-api-key" >}}).

Authenticated clients get **write-only** access to the following topics:

- `v3/<gateway-id>/up`: Used for sending uplink traffic to the Gateway Server.
- `v3/<gateway-id>/status`: Used for sending gateway status messages to the Gateway Server.
- `v3/<gateway-id>/down/ack`: Used for sending TxAck messages to the Gateway Server.

Clients also get **read-only** access and should subscribe to the following topic:

- `v3/<gateway-id>/down`: The Gateway Server publishes downlink message that the gateway should transmit.

## Disconnect from the Gateway Server

The gateway can disconnect by terminating the MQTT client connection.

## Uplink Messages

To forward uplink traffic to the Gateway Server, the MQTT client must publish a Protocol Buffer of type `ttnpb.UplinkMessage` under the topic `v3/<gateway-id>/up`.

Below is an example that connects to the Gateway Server as `$GATEWAY_ID`, publishes an uplink message (Protocol Buffer stored as binary file `test-uplink-message`) and disconnects:

```bash
$ export GATEWAY_ID="test-gtw"
$ export GATEWAY_API_KEY="NNSXS.VEEBURF3KR77ZR..." # API key with RIGHT_GATEWAY_LINK rights
$ mosquitto_pub \
    -h "thethings.example.com" -p 1882 \
    -u "$GATEWAY_ID" -P "$GATEWAY_API_KEY" \
    -t "v3/$GATEWAY_ID/up" -f test-uplink-message
```

>NOTE: Port 1882 is insecure. The TLS-enabled port 8882 should be used in a production setting.

>NOTE: The file `test-uplink-message` contains a Protocol Buffer of type `ttnpb.UplinkMessage` (binary, not JSON).

## Downlink Messages

The Gateway Server instructs the gateway to send a downlink packet by publishing a Protocol Buffer of type `ttnpb.GatewayDown` under the topic `v3/<gateway-id>/down`.

The MQTT client must subscribe to this topic after connecting to the Gateway Server. It must also listen for incoming `ttnpb.GatewayDown` messages (which contain both the packet data payload as well as any desired transmission settings). Upon receiving a scheduling request, it must trasmit that message, and [send back a `TxAck` packet]({{< ref "#txack-messages" >}}) on success.

```bash
$ export GATEWAY_ID="test-gtw"
$ export GATEWAY_API_KEY="NNSXS.VEEBURF3KR77ZR..." # API key with RIGHT_GATEWAY_LINK rights
$ mosquitto_sub \
    -h "thethings.example.com" -p 1882 \
    -u "$GATEWAY_ID" -P "$GATEWAY_API_KEY" \
    -t "v3/$GATEWAY_ID/down" -v
```

>NOTE: Port 1882 is insecure. The TLS-enabled port 8882 should be used in a production setting.

>NOTE: The example above is not complete (as it does nothing with the scheduled downlink requests). It is only meant to showcase the MQTT client subscribing to the downlink topic.

## Gateway Status Messages

To forward a gateway status message to the Gateway Server, the MQTT client must publish a Protocol Buffer of type `ttnpb.GatewayStatus` under the topic `v3/<gateway-id>/status`.

Below is an example that connects to the Gateway Server as `$GATEWAY_ID`, publishes a gateway status message (Protocol Buffer stored as binary file `test-gateway-status`) and disconnects:

```bash
$ export GATEWAY_ID="test-gtw"
$ export GATEWAY_API_KEY="NNSXS.VEEBURF3KR77ZR..." # API key with RIGHT_GATEWAY_LINK rights
$ mosquitto_pub \
    -h "thethings.example.com" -p 1882 \
    -u "$GATEWAY_ID" -P "$GATEWAY_API_KEY" \
    -t "v3/$GATEWAY_ID/status" -f test-gateway-status
```

>NOTE: Port 1882 is insecure. The TLS-enabled port 8882 should be used in a production setting.

>NOTE: The file `test-gateway-status` contains a Protocol Buffer of type `ttnpb.GatewayStatus` (binary, not JSON).

## TxAck Messages

To forward a `TxAck` packet to the Gateway Server, the MQTT client must publish a Protocol Buffer of type `ttnpb.TxAcknowledgement` under the topic `v3/<gateway-id>/down/ack`.

```bash
$ export GATEWAY_ID="test-gtw"
$ export GATEWAY_API_KEY="NNSXS.VEEBURF3KR77ZR..." # API key with RIGHT_GATEWAY_LINK rights
$ mosquitto_pub \
    -h "thethings.example.com" -p 1882 \
    -u "$GATEWAY_ID" -P "$GATEWAY_API_KEY" \
    -t "v3/$GATEWAY_ID/down/ack" -f example-tx-ack
```

>NOTE: Port 1882 is insecure. The TLS-enabled port 8882 should be used in a production setting.

>NOTE: The example above is not complete (the TxAck should be sent in response to a successful downlink transmission). It is only meant to showcase the MQTT client sending a TxAck packet (contents of the `example-tx-ack` file) to the Gateway Server.
