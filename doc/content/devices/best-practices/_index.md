---
title: "Best Practices"
description: ""
weight:
---

This section contains best practices for building connected devices which use {{% tts %}} as a Network Server.

<!--more-->

LoRaWAN devices should always comply to the [LoRaWAN specification](https://www.lora-alliance.org/lorawan-for-developers), with special regard to minimizing unnecessary join requests and duty cycle limitations.

The goal of these best practices is to optimize individual devices (especially for battery consumption) and the network as a whole (to reliably and efficiently serve more devices with the same number of gateways).

## Eliminate Unnecessary Join Requests

The LoRaWAN specification warns specifically against systematic rejoin in case of network failure. A device should keep the result of an activation in permanent storage if the device is expected to be power-cycled during its lifetime.

A device may temporarily lose connection with the network for many reasons: if the network server or gateways are suffering from an outage, if there’s no coverage in the area, etc. Whenever a device rejoins, it consumes the closest gateway’s airtime to emit the downlink - and if many devices in an area are rejoining at the same time (e.g. in case of a temporary gateway outage), this leads to network bloating in the area.

## Limit Transmission Length, Payload Size, and Duty Cycle

Limit the frequency of your transmissions to the minimum possible. Can you sample data once a day? Or even less?

Encode messages as efficiently as possible. Shorter messages mean shorter transmission time.

Avoid confirmed uplink messages. Confirmed uplinks should only be used in the case where 100% assurance of transmission is necessary, e.g. alarms.

The duty cycle of radio devices is often regulated by government. If this is the case, the duty cycle is commonly set to 1%, but make sure to check the regulations of your local government to be sure.

In Europe, duty cycles are regulated by section 7.2.3 of the ETSI EN300.220 standard.

Additionally, the LoRaWAN specification dictates duty cycles for the join frequencies, the frequencies devices of all LoRaWAN-compliant networks use for over-the-air activations (OTAA) of devices. In most regions this duty cycle is set to 1%.

## Expect Packet Loss

You should expect packet loss up to 10%. Implement Forward Error Correction if that's a problem.

## Synchronization, Backoff, and Jitter

([See LoRaWAN Specification 1.0.3, line 1065](https://lora-alliance.org/sites/default/files/2018-07/lorawan1.0.3.pdf)).

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

## Use a Good Random Number Generator

True randomness in a device's random number generator is especially important for preventing network congestion. If devices share a seed for a pseudorandom number generator, they will choose the same random numbers. Devices should use a unique seed such as the device address.

Bad randomization will result in an uneven distribution of channels selected for random channel selection, causing subpar network performance. ([LoRaWAN Specification 1.0.3, line 244](https://lora-alliance.org/sites/default/files/2018-07/lorawan1.0.3.pdf)).

Bad randomization will also result in transmission synchronization if devices respond to a large scale external event (for example, if they are all powered on at the same time).

## Use ADR for Stationary Devices

For devices that don't move, the LoRaWAN specification recommends allowing the Network Server to control the data rate to minimize power consumption.

For moving devices, ADR should not be used since RF conditions will likely change, but since many moving devices are temporarily stationary, it is possible to save additional power by requesting ADR only during the time a device is stationary. ([LoRaWAN Specification 1.0.3, line 438](https://lora-alliance.org/sites/default/files/2018-07/lorawan1.0.3.pdf)).

You may also use application specific knowledge to predict when ADR is appropriate. A tracking device can detect when it is moving, for example. A parked car sensor can detect when a parked car will affect RF conditions, and should fall back to another strategy. 

## Use OTAA

OTAA devices perform a join-procedure with the network, during which a dynamic Device Address is assigned and security keys are negotiated with the device. Activation by Personalization (ABP) requires hardcoding the Device Address as well as the security keys in the device, which is insecure. ABP also has the downside that devices can not switch network providers without manually changing keys in the device.

## Power Cycles

Devices should save network parameters between regular power cycles. This includes session parameters like `DevAddr`, session keys, `FCnt`, and nonces. This allows the device to easily Join, as keys and counters remain synchronized.

Devices should also randomize initial power on delay (i.e. Join). See [Synchronization, Backoff, and Jitter](#synchronization-backoff-jitter)

## Frame Counters

Devices must increment the frame counter after each uplink and downlink. Devices should use 32 bit counters for FCntUp and FCntDwn to prevent replay attacks.

## Ack

It is possible that you will not receive an ACK for every confirmed uplink or downlink. A good rule of thumb is to wait for at least **three** missed ACK's to assume link loss.

In the case of link loss, do the following:

- Set TX power to the maximum allowed/supported, and try again
- Decrease data rate step by step, and try again
- Reset to default channels, and try again
- Send periodic join requests with backoff
