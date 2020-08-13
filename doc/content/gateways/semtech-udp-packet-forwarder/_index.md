---
title: "Semtech UDP Packet Forwarder"
description: ""
weight: -1
---

The [{{% udp-pf %}}](https://github.com/lora-net/packet_forwarder) is the original LoRaWAN packet forwarder, connecting to servers through the Semtech UDP protocol. Many gateways include a pre-compiled version of the {{% udp-pf %}}, often adapted to the specific gateway.

The {{% udp-pf %}} has many security and scalability drawbacks, so if possible, use {{% lbs %}} to connect your gateway to {{% tts %}}. 

<!--more-->

## Configuration

When the packet forwarder starts, it looks in the current directory for a `global_conf.json`, a `local_conf.json` and a `debug_conf.json`. The Gateway EUI, Network Server Address, Frequency Plan, Ports, and other parameters are configurable in these files.

If `debug_conf.json` exists, the other files are ignored - otherwise, the parameters in `local_conf.json` override those in `global_conf.json`.

An example `global_conf.json` is available in the [{{% udp-pf %}} Github repository](https://github.com/Lora-net/packet_forwarder/blob/master/lora_pkt_fwd/global_conf.json). It is also possible to download a `global_conf.json` configured with your Gateway EUI and Frequency Plan directly from {{% tts %}}.

## Download Configuration in the Console

To download a `global_conf.json` file for your gateway, open the Gateway overview page in the console. Click the **Show global_conf.json** button to view the file, and the **Download** button to save it.

{{< figure src="conf.png" alt="Show global_conf.json" >}}

## Download Configuration via Terminal

To download a `global_conf.json` file using the terminal, you will need a Gateway API key with the `View gateway information` right enabled. To create an API key, see instructions for the [Console]({{< ref "/getting-started/console/create-gateway" >}}) or the [CLI]({{< ref "/getting-started/cli/create-gateway" >}}).

Open the command prompt in Windows or any Linux terminal to run a curl command (as shown below) to generate the required `global_conf.json` file in your current working directory.

Make sure you replace `thethings.example.com` with your server address, `{GATEWAY_ID}` with your Gateway EUI, and `{GATEWAY_API_KEY}` with the API key you generated:

```bash
$ curl -XGET \
    "https://thethings.example.com/api/v3/gcs/gateways/{GATEWAY_ID}/semtechudp/global_conf.json" \
    -H "Authorization: Bearer {GATEWAY_API_KEY}" > global_conf.json
```
