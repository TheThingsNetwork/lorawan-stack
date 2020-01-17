---
title: "Join Server"
description: ""
weight: 5
---

The Join Server handles the LoRaWAN join flow, including Network and Application Server authentication and session key generation.

<!--more-->

## Join procedure

Join Servers receive join-requests from [Network Servers]({{< ref "/components/network-server.md" >}}) via gRPC and issue join-accepts for registered devices if join-request validation passes.

In case a join-request is accepted, the Join Server derives session security context, which contains the session keys and is identified by a session key ID. Join Servers encrypt derived network and application session keys using key encryption keys(KEKs) shared between [Network Servers]({{< ref "/components/network-server.md" >}}) and [Application Servers]({{< ref "/components/application-server.md" >}}) respectively and include the session keys in the join-accepts in encrypted form.

## Device Management

Join Servers expose [JsEndDeviceRegistry]({{< ref "/reference/api/end_device#the-jsenddeviceregistry-service" >}}) service for end device management. Typical clients of this service are [Console]({{< ref "/components/console.md" >}}) and [CLI]({{< ref "/components/cli.md" >}}).

Join Servers store device root and session keys.

## Session Key Retrieval

Join Servers expose RPCs for retrieval of session keys given session key ID.

## Interoperability

Join Servers expose AS-JS, vNS-JS and hNS-JS services as defined by LoRaWAN Backend Interface 1.0 spec.
