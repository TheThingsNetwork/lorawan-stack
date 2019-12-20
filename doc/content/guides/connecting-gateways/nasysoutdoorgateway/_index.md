---
title: "NASys LoRaWAN Outdoor Gateway"
description: ""
weight: 1
---

NASys LoRaWAN Outdoor Gateway is an 8 Channel LoRaWAN gateway, whose technical specifications can be found in [the official product page](https://www.nasys.no/product/lorawan-gateway/). This page guides you to connect it to {{% tts %}}.

## Prerequisites

1. User account on {{% tts %}} with rights to create Gateways.
2. NASys LoRaWAN Outdoor Gateway connected to the internet (or your local network) via ethernet.

## Registration

Create a gateway by following the instructions for the [Console]({{< ref "/guides/getting-started/console#create-gateway" >}}) or the [CLI]({{< ref "/guides/getting-started/cli#create-gateway" >}}). Typically, the **EUI** field for your gateway should exist on the sticker at the bottom. Make note of the Gateway ID you choose, because it will be needed later.

## Configuration using a Terminal

Find the IP address the gateway. This can be done in various ways. You can connect your machine to the same local network as that of the gateway Ethernet connection and scan for open SSH ports or assign a static IP to the gateway and use that. Once the gateway IP address is found, ssh into it.

```bash
$ ssh root@<GatewayIP>
```

The default username is **root**, and the default password can be also found in the sticker.

Your gateway should come with a slightly modified version of the [Lora-net UDP packet forwarder](https://github.com/Lora-net/packet_forwarder) pre-installed at `/opt/nas-lgw`. There are two configuration files `global_conf.json` and `local_conf.json`, both located in `/opt/nas_lgw`.

The Gateway Configuration Server can be used to retrieve a proper `global_conf.json` configuration file for your gateway.

You will need a Gateway API key with the `View gateway information` right enabled. Instructions can be found in the relevant sections of the [Console]({{< ref "/guides/getting-started/console#create-gateway" >}}) or the [CLI]({{< ref "/guides/getting-started/cli#create-gateway" >}}) getting started guides.

Make sure to replace `thethings.example.com` with your server:

```bash
$ export GATEWAY_ID="<ID_OF_YOUR_GATEWAY_ON_TTS>"
$ export GTW_API_KEY="NNSXS.AAAAAAAAAAAAA.BBBBBBBBBBBBBBBBB"
$ curl -XGET \
    "https://thethings.example.com/api/v3/gcs/gateways/${GATEWAY_ID}/semtechudp/global_conf.json" \
    -H "Authorization: Bearer ${GTW_API_KEY}" > ~/global_conf.json
```

Then, update the configuration files and restart the packet forwarder:

```bash
$ mv /opt/nas-lgw/local_conf.json /opt/nas-lgw/local_conf.json.old
$ cp ~/global_conf.json /opt/nas-lgw/global_conf.json

$ systemctl restart nas_lgw
```

If your configuration was successful, your gateway will connect to {{% tts %}} after a couple of seconds.

## Troubleshooting

If the gateway does not connect to {{% tts %}} after a few minutes, issue a `reboot` command, or disconnect and reconnect the power supply to power-cycle the gateway.

If you still have trouble connecting to {{% tts %}}, then try editing the `gateway_conf` section:

```bash
$ vi /opt/nas-lgw/global_conf.json
```

Edit the server parameters:

1. **gateway_ID**: Make sure this is the same as the GatewayEUI (in lowercase).
2. **server_address**: Address of the Gateway Server. If you followed the [Getting Started guide]({{< ref "/guides/getting-started" >}}) this is the same as what you use instead of `thethings.example.com`.
3. **serv_port_up**: UDP upstream port of the Gateway Server, typically 1700.
4. **serv_port_down**: UDP downstream port of the Gateway Server, typically 1700.

You can access the gateway system logs using journalctl. See `journalctl --help` for details

```bash
$ journalctl -f -u nas_lgw -n 1000
```

> **IMPORTANT NOTE**: The gateway logs will rotate when they reach about 15M in size, which means that you will generally not be able to access very old logs. At times of dense traffic (e.g. ~1000s of devices) this typically means that you will only have logs for 2-3 hours. If you want to keep historical data (for whatever reason), then you will have to forward the logs to an external server. If you decide to do so, then `netcat` may be useful:

> ```bash
> $ journalctl -f | nc server-hostname server-port
> ```
