---
title: "Best Practices"
description: ""
weight:
---

This section contains best practices for building connected devices which use {{% tts %}} as a Network Server.

<!--more-->

LoRaWAN devices should always comply to the [LoRaWAN specification](https://www.lora-alliance.org/lorawan-for-developers), with special regard to minimizing unnecessary join requests and duty cycle limitations.

The goal of these best practices is to optimize individual devices (especially for battery consumption) and the network as a whole (to reliably and efficiently serve more devices with the same number of gateways).

## Unnecessary Join Requests

The LoRaWAN specification warns especially against systematic rejoin in case of network failure. A device should keep the result of an activation in permanent storage if the device is expected to be turned off during its lifetime.

A device may temporarily lose connection with the network for many reasons: if the network server or gateways are suffering from an outage, if there’s no coverage in the area, etc. Whenever a device rejoins, it consumes the closest gateway’s airtime to emit the downlink - and if many devices in an area are rejoining at the same time (e.g. in case of a temporary gateway outage), this leads to network bloating in the area.

## Limit Transmission Length, Payload Size, and Duty Cycle

Limit the frequency of your transmissions to the minimum possible. Can you sample data once a day? Or even less?

Encode messages as efficiently as possible. Shorter messages means shorter transmission time.

The duty cycle of radio devices is often regulated by government. If this is the case, the duty cycle is commonly set to 1%, but make sure to check the regulations of your local government to be sure.

In Europe, duty cycles are regulated by section 7.2.3 of the ETSI EN300.220 standard.

Additionally, the LoRaWAN specification dictates duty cycles for the join frequencies, the frequencies devices of all LoRaWAN-compliant networks use for over-the-air activations (OTAA) of devices. In most regions this duty cycle is set to 1%.

## Synchronization, Backoff, and Jitter

Synchronization of devices happens if end devices respond to a large-scale external event - for example, hundreds of end devices that are connected to the same power source and the power is switched off and on again, or hundreds of end devices that are connected to the same gateway, and the firmware of the gateway needs to be updated.

Let’s take an example device that starts in Join mode when it powers on, and reverts to Join mode after being disconnected from the network. There are hundreds of such devices in a field, and one gateway that covers this field.

The power source for the devices is switched on, and the gateway immediately receives the noise of hundreds of simultaneous Join requests. LoRaWAN gateways can deal quite well with noise, but this is just too much, and the gateway can’t make any sense of it. No Join requests are decoded, so no Join requests are forwarded to the network and no Join requests are accepted.

If the retry interval is a fixed duration, e.g. 10 seconds, the gateway again receives the noise of hundreds of simultaneous Join requests, and still can’t make anything of it. This continues every 10 seconds after that, and the entire site stays offline.

### Jitter

This situation can be improved by using jitter. Instead of sending a Join request every 10 seconds, the devices send a Join request 10 seconds after the previous one, plus or minus a random duration of 0-20% of this 10 seconds. This jitter percentage needs to be truly random, because if all devices use the same pseudorandom number generator, they will still be synchronized, as they will all pick the same “random” number.

With these improved devices, the Join requests will no longer all be sent at exactly the same time, and the gateway will have a better chance of decoding the Join requests.

### Backoff

But what if you have another site with thousands of these devices? Then the 10 seconds between Join messages may not be enough. This is where backoff comes in. Instead of having a delay of 10s±20%, you increase the delay after each attempt, so you do the second attempt after 20s±20%, the third after 30s±20%, and you keep increasing the delay until you have, say, 1h±20% between Join requests.

An implementation like this prevents persistent failures of sites and the network as a whole and helps speed up recovery after outages.

## Additional Best Practices

- Save device parameters between regular power cycles
- Use a true random number generator (especially for jitter and channel hopping)
- Use [OTAA]({{< ref "/reference/glossary#otaa" >}}) instead of [ABP]({{< ref "/reference/glossary#abp" >}})
- Optimize [data rate]({{< ref "/reference/glossary#data-rate" >}})
- Try to use [ADR]({{< ref "/reference/glossary#adr" >}}) for non moving devices, and ADR for moving devices
- Avoid non-essential downlinks
