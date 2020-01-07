---
title: "Kerlink Wirnet Station"
description: ""
weight: 1
---

Kerlink Wirnet Station is a LoRaWAN gateway, whose technical specifications can be found in [the official documentation](https://www.kerlink.com/product/wirnet-station/). This page guides you to connect it to {{% tts %}}.

## Prerequisites

1. User account on {{% tts %}} with rights to create Gateways.
2. Kerlink Wirnet Station with Common Packet Forwarder installed and enabled and/or running firmware version `3.0` or higher.

> NOTE: Minimum CPF version tested is `v1.1.6`

## Registration

Create a gateway by following the instructions for the [Console]({{< ref "/guides/getting-started/console#create-gateway" >}}) or the [CLI]({{< ref "/guides/getting-started/cli#create-gateway" >}}). Choose a **Gateway ID** and set **EUI** equal to the one on the gateway.

Create an API Key with Gateway Info rights for this gateway using the same instructions. Copy the key and save it for later use.

## Configuration

All further steps will assume the gateway is available at `192.168.4.155`, the stack address is `thethings.example.com`, gateway ID is `example-gtw` and gateway API key is `NNSXS.GTSZYGHE4NBR4XJZHJWEEMLXWYIFHEYZ4WR7UAI.YAT3OFLWLUVGQ45YYXSNS7HTVTFALWYSXK6YLJ6BDUNBPJMRH3UQ`, please replace these by the values appropriate for your setup.

### Provisioning

1. Execute: 
```bash
$ curl -sL https://raw.githubusercontent.com/TheThingsNetwork/kerlink-wirnet-firmware/v0.0.1/provision.sh | bash -s -- 'wirnet-station' '192.168.4.155' 'thethings.example.com' 'example-gtw' 'NNSXS.GTSZYGHE4NBR4XJZHJWEEMLXWYIFHEYZ4WR7UAI.YAT3OFLWLUVGQ45YYXSNS7HTVTFALWYSXK6YLJ6BDUNBPJMRH3UQ'
```

Please refer to [Kerlink Wirnet provisioning documentation](https://github.com/TheThingsNetwork/kerlink-wirnet-firmware/tree/v0.0.1#provisioning) if more detailed up-to-date documentation is necessary.

> NOTE: To avoid being prompted for `root` user password several times, you may add your SSH public key as authorized for `root` user on the gateway, for example, by `ssh-copy-id root@192.168.4.155`.

## Troubleshoting

Packet forwarder logs are located at `/mnt/fsuser-1/lora/var/log/lora.log`. You can access them by e.g.:

```bash
ssh root@192.168.4.155 'tail -f /mnt/fsuser-1/lora/var/log/lora.log'
```
