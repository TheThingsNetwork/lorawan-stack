---
title: "Eclipse Mosquitto"
description: ""
---

[Eclipse Mosquitto](https://mosquitto.org/) is a project which provides an open source MQTT broker, a C and C++ library for MQTT client implementations and the popular command line MQTT clients. Its lightweight MQTT protocol implementation makes it suitable for full power machines, as well as for the low power and embedded ones. 

<!--more-->

This guide shows how to receive upstream messages and send downlink messages with the Eclipse Mosquitto command line clients and {{% tts %}} [MQTT Server]({{< ref "/integrations/mqtt" >}}).

>Note: Eclipse Mosquitto MQTT server supports 3.1, 3.1.1 and 5.0 MQTT protocol versions.

## Requirements

1. [Eclipse Mosquitto MQTT server](https://github.com/eclipse/mosquitto) installed on your system.

## Subscribing to Upstream Traffic

>Note: this section follows the example for subscribing to upstream traffic in the [MQTT Server]({{< ref "/integrations/mqtt#subscribing-to-upstream-traffic" >}}) guide.

The command for connecting to a host and subscribing to a topic has using `mosquitto_sub` has the following syntax:

```bash 
mosquitto_sub -h {hostname} -p {port} -u {username} -P {password} -t {topic}
```

For example, to subscribe to all topics in the application `app1`:

```bash
# Tip: when using `mosquitto_sub`, pass the `-d` flag to see the topics messages get published on.
# For example:
$ mosquitto_sub -h thethings.example.com -t "#" -u app1 -P "NNSXS.VEEBURF3KR77ZR.." -d
```

In you want to use TLS, you need to change the port value to `8883` and add the `--cafile` option to the command. `--cafile` option is used to define a path to the file containing trusted CA certificates that are PEM encoded.

>Read more about the command line options in the [mosquitto_sub manual](https://mosquitto.org/man/mosquitto_sub-1.html).

## Publishing Downlink Messages

>Note: this section follows the example for publishing downlink traffic in the [MQTT Server]({{< ref "/integrations/mqtt" >}}) guide. See [Publishing Downlink Messages]({{< ref "/integrations/mqtt#publishing-downlink-traffic" >}}) for a list of available topics.

For connecting to a host and publishing a message, **mosquitto_pub** client defines a command with the following syntax:

```bash 
mosquitto_pub -h {hostname} -p {port} -u {username} -P {password} -t {topic} -m {message}
```

For example, to send an unconfirmed downlink message to the device `dev1` in application `app1` with the hexadecimal payload `BE EF` on `FPort` 15 with normal priority, use the topic `v3/app1/devices/dev1/down/push` with the following contents:

```bash
mosquitto_pub -h "thethings.example.com" -p "1883" -u "app1" -P "NNSXS.VEEBURF3KR77ZR.." -t "v3/app1/devices/dev1/up" -m '{"downlinks":[{"f_port": 15,"frm_payload":"vu8=","priority": "NORMAL"}]}'
```

If TLS is being used, change the port value to `8883` and add the `--cafile` option to the command.

>Read more about the command line options in the [mosquitto_pub manual](https://mosquitto.org/man/mosquitto_pub-1.html).
