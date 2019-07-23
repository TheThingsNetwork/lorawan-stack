---
title: "Downlink Queue Operations"
description: ""
weight: 30
---

{{< cli-only >}}

The stack keeps a queue of downlink messages per device. Applications can keep pushing downlink messages or replace the queue with a list of downlink messages.

If there are more application downlink messages in the queue, the Network Server sets the LoRaWAN `FPending` bit to indicate end devices that there is more downlink available. In class A downlink, this typically triggers the device to send an uplink message to receive the downlink message. In class C, the Network Server automatically transmits all queued downlink messages.

You can schedule downlink using the CLI, [MQTT server]({{< relref "../getting-started/mqtt" >}}) or [HTTP webhooks]({{< relref "../getting-started/webhooks" >}}).

## Push and replace downlink queue

To push downlink to the end of the queue:

```bash
$ ttn-lw-cli end-devices downlink push app1 dev1 \
  --frm-payload 01020304 \
  --priority NORMAL
```

You can pass an `FPort` (default `1`) with `--f-port`, and confirmed downlink with `--confirmed`.

To replace the existing queue with a new item:

```bash
$ ttn-lw-cli end-devices downlink replace app1 dev1 \
  --frm-payload 01020304 \
  --priority NORMAL
```

## List queue

To see currently scheduled downlink messages:

```bash
$ ttn-lw-cli end-devices downlink list app1 dev1
```

## Clear queue

To clear scheduled downlink messages:

```bash
$ ttn-lw-cli end-devices downlink clear app1 dev1
```
