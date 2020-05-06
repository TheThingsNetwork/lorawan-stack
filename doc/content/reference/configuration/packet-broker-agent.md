---
title: "Packet Broker Agent Options"
description: ""
---

## Connection Options

- `pba.data-plane-address`: Address of Packet Broker Data Plane
- `pba.net-id`: LoRa Alliance NetID
- `pba.tenant-id`: Tenant ID within the NetID
- `pba.cluster-id`: Cluster ID uniquely identifying this cluster within a NetID and tenant. The cluster ID is used for shared subscriptions (i.e. splitting traffic over multiple Packet Broker Agents) and as Forwarder ID to route downlink traffic to the right cluster

## Client TLS Options

Packet Broker Agent uses TLS client authentication to connect to the configured Packet Broker Data Plane. You need to configure a client certificate which is authorized for the configured NetID and tenant ID.

- `pba.tls.source`: Source of the TLS certificate (`file`, `key-vault`)

If `file` is specified as `pba.tls.source`, the location of the certificate and key need to be configured.

- `pba.tls.certificate`: Location of TLS certificate
- `pba.tls.key`: Location of TLS private key

If `key-vault` is specified as `pba.tls.source`, the certificate with the given ID is loaded from the key vault.

- `pba.tls.key-vault.id`: ID of the certificate

## Forwarder Options

- `pba.forwarder.enable`: Enable Forwarder role
- `pba.forwarder.worker-pool.limit`: Limit of active workers concurrently forwarding uplink messages and processing downlink messages
- `pba.forwarder.token-key`: AES 128 or 256-bit key for encrypting uplink tokens

## Home Network Options

- `pba.home-network.enable`: Enable Home Network role
- `pba.home-network.dev-addr-prefixes`: DevAddr prefixes to subscribe to
- `pba.home-network.worker-pool.limit`: Limit of active workers concurrently processing uplink messages and publishing downlink messages
- `pba.home-network.blacklist-forwarder`: Blacklist traffic from Forwarder to avoid traffic loops. Enable this when you have the Forwarder role enabled and overlapping forwarding ranges in `gs.forward` (see [configuration]({{< relref "gateway-server.md" >}})). Only disable this for testing.
