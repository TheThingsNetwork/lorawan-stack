---
title: "HiveMQ"
description: ""
weight: 
---

[HiveMQ](https://www.hivemq.com/) is an MQTT broker and a client based messaging platform which uses MQTT protocol for fast, reliable and efficient bi-directional data transfer to and from IoT devices. HiveMQ provides its own client library, but it can be used with any MQTT compliant client library. It can be deployed on a private, hybrid or public cloud. You can integrate HiveMQ with existing enterprise systems thanks to its open API and a flexible extension framework.

<!--more-->

HiveMQ also offers an open source tool called [MQTT CLI](https://github.com/hivemq/mqtt-cli), which provides a command line interface to interact with MQTT brokers. This tool can be used in a shell mode, allowing you to use multiple MQTT clients simultaneously. 

>Note: HiveMQ broker is compliant with the MQTT 3.1, 3.1.1 and 5.0 protocol specifications, while the MQTT CLI tool supports 3.1.1 and 5.0 versions.

This guide contains the instructions to use HiveMQ CLI tool in a shell mode for subscribing and publishing to topics used by {{% tts %}} [MQTT Server]({{< ref "/integrations/mqtt" >}}).

## Requirements

1. [HiveMQ MQTT CLI](https://hivemq.github.io/mqtt-cli/docs/installation.html) installed on your system.

## Connecting to MQTT Server in MQTT CLI Shell Mode

>Learn how to connect to {{% tts %}} MQTT Server by reading the [MQTT Server]({{< ref "/integrations/mqtt" >}}) guide.

Enter the HiveMQ MQTT CLI shell mode by typing the following command in your terminal:

```bash
$ mqtt shell
```

Once in shell mode, you can connect to {{% tts %}} MQTT Server by using the following command:

```bash
$ con -h hostname -p port -V 3 -u username -pw password
```

>Note: keep in mind that `password` is the value of the authentication API key. For more info, see [Creating an API Key]({{< ref "/integrations/mqtt#creating-an-api-key" >}}).

>Note: we use `-V 3` flag since the {{% tts %}} MQTT Server supports the 3.1.1 MQTT protocol version, as mentioned in the [MQTT Server guide]({{< ref "/integrations/mqtt" >}}). For detailed descriptions of other parameters used with the `con` command, see the [official MQTT CLI documentation](https://hivemq.github.io/mqtt-cli/docs/shell/connect.html).

For example, you can connect to {{% tts %}} MQTT Server over its public address with the following command:

```bash
con -h thethings.example.com -p 1883 -V 3 -u app1 -pw NNSXS.VEEBURF3KR77ZR..
```

TO use TLS for additional security, change the port from `1883` to `8883` and use the `--cafile` option to provide the PEM encoded CA file of your {{% tts %}} deployment.

Once you have successfully connected to {{% tts %}} MQTT Server, continue with subscribing or publishing to topics exposed by it by following the sections below.

## Subscribe to Upstream Traffic

Use the `sub` command to subscribe to topics and listen to messages being sent from your end device. 

For example, if you want to listen to the uplink messages being sent from `dev1` device in `app1` application, use the following command:

```bash
$ sub -t v3/app1/devices/dev1/up -s
```

>Note: `-s` flag is used to subscribe with a context to the given topic, e.g. to stop the console being blocked by subscribing without a context. For detailed descriptions of all the available `sub` command parameters, see the [Subscribe](https://hivemq.github.io/mqtt-cli/docs/shell/subscribe.html) section of the HiveMQ MQTT CLI documentation.

>See the [Subscribing to Upstream Traffic]({{< ref "/integrations/mqtt#subscribing-to-upstream-traffic" >}}) section of the MQTT Server guide for a full list of available topics you can subscribe to. 

## Schedule Downlink Messages

Use the `pub` command to publish to topics, e.g. to schedule downlink messages to be sent to your end device. 

For example, to push an unconfirmed downlink message with the hexadecimal payload `BE EF` on `FPort` 15 with normal priority to the `dev1` device, use the following command:

```bash
$ pub -t v3/app1/devices/dev1/down/push -m '{"downlinks":[{"f_port": 15,"frm_payload":"vu8=","priority": "NORMAL"}]}'
```
>Note: for detailed descriptions of the `pub` command parameters, see the [Publish](https://hivemq.github.io/mqtt-cli/docs/shell/publish.html) section of the MQTT CLI documentation.

>See the [Publishing Downlink Traffic]({{< ref "/integrations/mqtt#publishing-downlink-traffic" >}}) section to learn about using `/replace` instead of `/push`.