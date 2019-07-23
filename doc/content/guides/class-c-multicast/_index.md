---
title: "Class C and Multicast"
description: ""
weight: 40
---

In order to send Class C downlink messages to a single device, enable Class C support for the end device when creating or updating it with the `--supports-class-c` flag.

For example, when enabling Class C for an existing device:

```bash
$ ttn-lw-cli end-devices update app1 dev1 --supports-class-c
```

This will enable the Class C downlink scheduling of the device. That's it! Downlink messages are now scheduled as soon as possible.

To disable Class C scheduling, reset with `--supports-class-c=false`.

>Note: Class C downlink scheduling starts when the end device confirms the session. This means that the device should send an uplink message after receiving the join-accept in order to enable Class C downlink scheduling.

## Class C message settings

TTN Stack supports optional settings for Class C downlink messages: the downlink path and the time to send the message.

The downlink path is defined by one or more gateways IDs. The Network Server and Gateway Server schedules only on the specified gateways in the specified order. This is useful for multicast (where no downlink path is known because there is no uplink). A scheduling attempt on can fail when the gateway is not connected, if there is a scheduling conflict or if duty-cycle regulations prohibit transmission. See the [Example]({{< relref "#example" >}}) below.

The time to transmit is an absolute timestamp in ISO 8601 format to send the message. This requires gateways either with GPS lock, or gateways that use a protocol that provide round-trip times (RTT). See the [Example]({{< relref "#example" >}}) below.

## Multicast group

It is also possible to create a multicast group to send a Class C downlink message to a group of end devices. A multicast group is a virtual ABP device (i.e. shared session keys), does not support uplink, confirmed downlink nor MAC commands.

When creating a device, you can specify in the Console and CLI whether it's a multicast group.

<details><summary>Show CLI example</summary>
```bash
$ ttn-lw-cli end-devices create app1 mc1 \
  --frequency-plan-id EU_863_870 \
  --lorawan-version 1.0.3 \
  --lorawan-phy-version 1.0.3-b \
  --session.dev-addr 00E4304D \
  --session.keys.app-s-key.key A0CAD5A30036DBE03096EB67CA975BAA \
  --session.keys.nwk-s-key.key B7F3E161BC9D4388E6C788A0C547F255 \
  --multicast
```
</details>

>Note: A multicast group cannot be converted to a normal unicast device or the other way around.

>Note: Since multicast does not support uplink, the Network Server does not know a downlink path. Therefore, you need to specify a downlink path when scheduling downlink message.

## Example

{{< cli-only >}}

First, create a multicast group:

```bash
$ ttn-lw-cli end-devices create app1 mc1 \
  --frequency-plan-id EU_863_870 \
  --lorawan-version 1.0.3 \
  --lorawan-phy-version 1.0.3-b \
  --session.dev-addr 00E4304D \
  --session.keys.app-s-key.key A0CAD5A30036DBE03096EB67CA975BAA \
  --session.keys.nwk-s-key.key B7F3E161BC9D4388E6C788A0C547F255 \
  --multicast
```

Then, schedule the following message to the [MQTT server]({{< relref "../getting-started/mqtt" >}}) or [HTTP webhooks]({{< relref "../getting-started/webhooks" >}}):

```json
{
  "downlinks": [{
    "frm_payload": "vu8=",
    "f_port": 42,
    "priority": "NORMAL",
    "class_b_c": {
      "gateways": [
        {
          "gateway_ids": {
            "gateway_id": "gtw1"
          },
        },
        {
          "gateway_ids": {
            "gateway_id": "gtw2"
          },
        }
      ],
      "absolute_time": "2019-07-23T13:05:00Z"
    }
  }]
}
```
