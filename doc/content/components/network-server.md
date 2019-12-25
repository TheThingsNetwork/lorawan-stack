---
title: "Network Server"
description: ""
weight: 3
---

The Network Server handles the LoRaWAN network layer, including MAC commands, regional parameters and adaptive data rate (ADR).

<!--more-->

## Device Management

Network Servers expose [NsEndDeviceRegistry]({{< ref "/reference/api/end_device#the-nsenddeviceregistry-service" >}}) service for end device management. Typical clients of this service are [Console]({{< ref "/components/console.md" >}}) and [CLI]({{< ref "/components/cli.md" >}}).

Network Servers store device MAC configuration, MAC state and network session keys.

Change of device MAC configuration may trigger a downlink message.

## Application Downlink Queue Management and Linking

Network Servers let [Application Servers]({{< ref "/components/gateway-server.md" >}}) push, replace and list application downlinks as well as link applications via gRPC API.

Change of application downlink queue may trigger a downlink message.

Once the link is established, Network Server will send application-specific uplink messages to the client via the link. There can be at most one active link per-appplication.

In case link is not active, but Network Server has application-specific uplink messages to send, those messages will be queued and sent once the link is established.

## Downlink Scheduling

Network Servers maintain internal downlink task queue. Each downlink task has an execution time associated with it, which tasks are sorted by in ascending order. Whenever a downlink task is ready to execute, it is executed as soon as possible.

### Join-accept

In case a pending session exists and join-accept is queued for the device, it is scheduled.

### Data downlink

In case a pending session does not exist or join-accept for it has already been sent, Network Server attempts to genererate and schedule data downlink in the active session.

## Uplink Handling

Network Servers receive uplinks from [Gateway Servers]({{< ref "/components/gateway-server.md" >}}) via gRPC.

Network Servers process the uplinks received and handle accordingly. The first step is matching the uplink to a device. In case an uplink cannot be matched to a device stored in Network Server, it is dropped.

### Join-request

If join-request is received:

1. Device is matched using the `DevEUI` and `JoinEUI` pair present in the join-request, which uniquely identifies the device.
2. New `DevAddr` is assigned to the device and new MAC state is derived for the device.
3. If [Join Server]({{< ref "/components/join-server.md" >}}) is present in the cluster, Network Server sends a join-request message to the cluster-local [Join Server]({{< ref "/components/join-server.md" >}}).
4. If [Join Server]({{< ref "/components/join-server.md" >}}) is not present in the cluster or the device is not provisioned in the cluster-local [Join Server]({{< ref "/components/join-server.md" >}}), Network Server sends a join-request message to the [Join Server]({{< ref "/components/join-server.md" >}}) discovered via [interoperability configuration]({{< ref "/reference/interop-repository" >}}).
5. If a [Join Server]({{< ref "/components/join-server.md" >}}) accepted the join-request, join-accept message may be enqueued for the device and application-specific uplink messages carrying relevant information about join-accept is sent to linked [Application Server]({{< ref "/components/application-server.md" >}}).

### Data uplink

If data uplink is received:

1. Device is matched using the `DevAddr` present in the data uplink. Matching is performed by comparing the session context and MAC state, as well as performing the MIC check. Since several devices may have identical `DevAddr`s, the Network Server may need to go through several stored devices before matching the device.
2. The Network Server processes the MAC commands, if such are present in the frame and updates MAC state accordingly.
3. If ADR bit is set in data uplink, Network Server runs the ADR algorithm and updates MAC state accordingly.
4. If data uplink is successfully processed, downlink may be enqueued for the device and one or more application-specific uplink messages carrying relevant information about join-accept are sent to linked [Application Server]({{< ref "/components/application-server.md" >}}).
