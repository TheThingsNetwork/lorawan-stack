---
title: "Troubleshooting"
description: ""
---

This section provides help for common issues and frequently asked questions you may have when getting started with {{% tts %}}. 

<!--more-->

If you encounter a bug, please [file an issue](https://github.com/TheThingsNetwork/lorawan-stack/issues/new/choose). If you have questions not covered here, please ask in [{{% tts %}} Forum](https://www.thethingsnetwork.org/forum/c/network-and-routing/v3). To check the status of cloud services, see our [status page](https://status.thethings.industries/).

## Troubleshooting {{% tts %}}

### How do I generate and authorize OAuth client keys?

See the [authentication reference]({{< ref src="reference/api/authentication" >}}).

### How do I create and authorize new users?

See [user and organization management]({{< ref src="getting-started/user-management" >}}).

## Troubleshooting Devices

### How do I see device events?

Device event logs can be found in the console in the device's general information page. See [Working with Events]({{< ref src="getting-started/events" >}}) for other ways of subscribing to events.

### My device won't join. What do I do?

- Double check your DevEUI, JoinEUI or AppEUI, LoRaWAN and Regional Parameters Version, root keys (AppKey, and with LoRaWAN 1.1 or higher, NwkKey)
- Check gateway and application events for traffic from your device
- Double check frequency plan settings in your end device and gateways (they must be the same LoRaWAN band)
- Double check your network connection. If there is a slow connection from the server to the gateway, the join accept message may be sent too late (this can happen when a gateway uses 3G as a backhaul). If using the CLI, run `ttn-lw-cli gateways connection-stats <gateway-id>` to see the round trip time (RTT) for your gateway
- Check for duplicate use of JoinNonce (or AppNonce) and/or DevNonce

### No downlinks are reaching my device. What do I do?

- Check gateway and application events for traffic from your device
- Check duty cycle restrictions
- Device clock drift often occurs when SF12 is used
- Check your antenna connections

### {{% tts %}} is no longer receiving uplinks from my device. What do I do?

- Check [gateway events](#how-do-i-see-gateway-events) and [device events](#how-do-i-see-device-events) for traffic from your device
- Check for FCnt mismatch on ABP devices

## Troubleshooting Gateways

### How do I see gateway events?

Gateway event logs can be found in the console in the gateway's general information page. See [Working with Events]({{< ref src="getting-started/events" >}}) for other ways of subscribing to events.

### My gateway won't connect. What do I do?

- Double check your gateway settings in {{% tts %}}. Ensure your Gateway EUI is correct
- Double check the configuration settings on your gateway. Is the network address the same address you use to connect to {{% tts %}}? Are your ports correct?
- Check your gateway logs. For information specific to your gateway, see the [Gateways section]({{< ref src="/gateways" >}})
- If using Basic Station with the CUPS protocol enabled, double check your API Keys and certificates

## Troubleshooting Common Errors

### Fail Link / Network Server Peer Not Available

This happens when an Application is not linked properly to the Network Server. {{% tts %}} does not support external linking of the Application Server. Ensure that in your application, the Network Server address is blank and that the application API key is valid with linking rights.

### Entity Not Found / Device Not Found

Your gateway receives traffic from all devices in range, and {{% tts %}} drops this traffic if no matching device is found. If you are trying to activate a device but receiving this error, double check your DevEUI and JoinEUI (or AppEUI).

In rare cases, this may also happen to ABP devices if a device resets its frame counters (usually because of a power cycle) and can not be found by the session. In this case, it is necessary to re-activate the device. OTAA prevents this problem from occurring, and is more secure in general, so it is the preferred form of activation when possible.

### Duplicate Uplink

An uplink was received and forwarded by multiple gateways, but due to backhaul latency or a Packet Broker one arrived outside the deduplication window and was dropped.

### Invalid Major / JoinRequestPHYPayload Length / Unknown MType

These are typically non-LoRaWAN traffic received by a gateway on your network.
