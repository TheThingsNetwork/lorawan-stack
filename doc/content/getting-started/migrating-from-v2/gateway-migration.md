---
title: Gateway Migration
weight: 55
---

Next up is migrating gateways from {{% ttnv2 %}} to {{% tts %}}.

For instructions on adding gateways to {{% tts %}} using the CLI or Console, see [Adding Gateways]({{< ref "gateways/adding-gateways" >}}).

When using the Semtech UDP Packet Forwarder, make sure to update the `server_address` in the gateway configuration settings to the address of the Gateway Server (e.g. `my-tts-network.nam1.cloud.thethings.industries`).

When using the Semtech UDP Packet Forwarder, make sure to update the `server address` in the gateway configuration settings to the address of the Gateway Server (e.g. `my-tts-network.nam1.cloud.thethings.industries`).

Once your gateways are migrated, data will be routed to {{% tts %}}.

>Note: If you are within range of The Things Network, data might still end up in {{% ttnv2 %}}. If this occurs, consider disabling the devices in {{% ttnv2 %}} by deleting session keys, or completely deleting the application. 