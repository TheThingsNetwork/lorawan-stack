---
title: Enable Peering
description: ""
---

Peering is exchanging LoRaWAN traffic with other networks to share coverage and improve the overall network performance. See [Peering]({{< ref "/concepts/peering" >}}) for more information about peering.

This guide shows you how to enable peering on your private LoRaWAN network.

<!--more-->

## Prerequisites

1. A LoRa Alliance NetID or a tenant of a host NetID. To obtain a NetID, [become a member of the LoRa Alliance](https://lora-alliance.org/become-a-member)
2. A TLS client certificate to authenticate with your NetID (and tenant ID). [Learn how to obtain a TLS client certificate](https://github.com/packetbroker/pb/tree/master/configs)
3. If you are a tenant of a host NetID, your host must have configured DevAddr prefixes for your tenant
4. {{% tts %}} installed and configured. See [Getting Started]({{< ref "/guides/getting-started" >}})
5. Packet Broker CLI installed and configured. See [Packet Broker CLI](https://github.com/packetbroker/pb)

## Define DevAddr Prefix by NetID

The NetID is a 24 bit number issued by the LoRa Alliance. NetIDs are used in device addresses (DevAddr), so that networks and data exchanges know which network is serving the device.

Enter your NetID to obtain your DevAddr prefix:

{{< dev-addr-prefix >}}

Your DevAddr prefix is: <code><span data-content="dev-addr-prefix"></span></code>

>This guide uses The Things Network NetID `000013` as example, which has DevAddr prefix `26000000/7`.

## Configure Packet Broker Agent

The Packet Broker Agent component of {{% tts %}} connects to Packet Broker. The Packet Broker Agent can be configured as forwarder and as home network in your `ttn-lw-stack.yaml` configuration file:

```yaml
# Add Packet Broker configuration to your configuration file:

# Packet Broker Agent configuration
pba:
  # See https://packetbroker.org for available hosts
  data-plane-address: 'eu.packetbroker.io'
  net-id: '000013'
  tenant-id: 'demo' # Leave empty if you own the NetID and you don't use tenants
  cluster-id: 'demo'
  tls:
    source: 'file'
    certificate: 'pb.pem'
    key: 'pb-key.pem'
  forwarder:
    enable: 'true'
    # generate 16 bytes (openssl rand -hex 16)
    token-key: '00112233445566770011223344556677'
  home-network:
    enable: 'true'
    dev-addr-prefixes:
    - '26000000/7'
    blacklist-forwarder: 'false' # Important: set to true in production environments
```

See [Packet Broker Agent configuration]({{< ref "/reference/configuration/packet-broker-agent" >}}) for all configuration options.

## Configure Gateway Server

Configure Gateway Server to forward traffic for the current network to the Network Server in the cluster, and route all traffic to Packet Broker (via Packet Broker Agent):

```yaml
# Edit the Gateway Server configuration in your configuration file:

# Gateway Server configuration
gs:
  forward:
  # Forward traffic to the Network Server in the cluster
  - 'cluster=26000000/7'
  # Forward all traffic also to Packet Broker
  - 'packetbroker=00000000/0'
```

See [Gateway Server configuration]({{< ref "/reference/configuration/gateway-server" >}}) for all configuration options.

## Configure Network Server

Configure Network Server to issue device addresses (DevAddr) that fall within your NetID:

```yaml
# Edit the Network Server configuration in your configuration file:

# Network Server configuration.
ns:
  net-id: '000013'
```

If you are dividing your DevAddr prefix in smaller blocks, you can configure the Network Server to use specific DevAddr prefixes:

```yaml
# Network Server configuration.
ns:
  net-id: '000013'
  dev-addr-prefixes:
  - '26010000/16'
  - '26020000/16'
```

>By default, Network Server uses NetID `000000` which is intended for experimentation purposes. Only devices that are activated with a DevAddr that falls in your NetID will have their traffic routed by Packet Broker to your network.

See [Network Server configuration]({{< ref "/reference/configuration/network-server" >}}) for all configuration options.

## Test Uplink and Downlink

See [Publish and Subscribe Traffic](https://github.com/packetbroker/pb#publish-and-subscribe-traffic) on how to publish test messages and subscribe to traffic using Packet Broker CLI.

>Packet Broker Agent uses the configured `cluster-id` as Forwarder ID and subscription group.
