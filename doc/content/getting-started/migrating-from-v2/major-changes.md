---
title: Major Changes in V3
weight: 10
---

Before getting started, we will discuss major breaking changes between
the public community network (v2) and {{% tts %}}, along with some guidelines to make the
migration process easier to manage.

### Application Data

One of the first changes that you must handle when migrating to {{% tts %}} is
the new data format for uplink and downlink traffic.

Example uplink message on the V2:

```js
/* topic name: app1/devices/dev1/up */
{"app_id":"app1","dev_id":"dev1","hardware_serial":"1122334411223344","port":1,"counter":0,"payload_raw":"EQ==","payload_fields":{"led":17},"metadata":{"time":"2020-05-01T00:04:41.258830149Z","latitude":47.984,"longitude":43.123,"altitude":100}}
```

Example downlink message on the V2:

```js
/* topic name: app1/devices/dev1/down */
{"port":10,"confirmed":true,"payload_raw":"EBA="}
```

The data format has changed in {{% tts %}}. It uses a different schema, uses
different names and has much richer metadata support. You can read more about
that in [Data Formats]({{% ref "/integrations/data-formats" %}}) and in
[Working With Data]({{% ref "/getting-started/working-with-data" %}}).

One of the first things that you will have to do is to make sure that your
application can propely handle the new V3 data format.

### Payload formats

The public community network (v2) has support for payload decoders, converters,
validators (for uplink) and encoders (for downlink). These can be either CayenneLP
or Javascript functions.

In {{% tts %}} (v3), you can have an uplink payload formatter (similar to the
payload decoder) and a downlink payload formatter (similar to the payload
encoder). These can be set per application, and even be overridden per end device.
There are no converters and/or validators yet.

Migrating the payload decoder (v2) to an uplink payload formatter should be
straightforward, since they have the same format. Same goes for the payload
encoder (v2).

### LoraWAN support

{{% tts %}} requires the LoraWAN MAC Protocol Version and Regional Parameters
(LoraWAN PHY version) to be set per device. These default to `MAC_V1_0_2` and
`MAC_PHY_V1_0_2_REV_B` for devices imported from the Public Community Network,
in order to stay as compatible as possible.

The LoraWAN MAC settings are configurable per device instead of per application.

### MQTT Traffic

You will need to change the MQTT server your application connects to. This
needs to change for the MQTT address of the public community network to the
MQTT address of {{% tts %}}. You will also need to create API keys and updating
your MQTT credentials accordingly.

Read more about MQTT traffic in [MQTT Server]({{ ref "/integrations/mqtt" }}).

{{% tts %}} also supports HTTP webhooks and PubSubs for nats.io or mqtt.

### Storage Integration

{{% tts %}} does not currently support a Storage integration similar to the one
on the public community network. This feature will be added in a future release.

### Gateway coverage

The Packet Broker enables peering between networks, so traffic received by one
network (e.g. the public community network) but intended for a different
network ({{% tts %}}) can be forwarded to and from that network.

With Packet Broker enabled (which is not discussed here), you can receive
traffic on {{% tts %}} without the need to configure any of your gateways (since
traffic will be automatically forwarded from the public community network).

For private {{% tts %}} deployments with Packet Broker disabled, you will need
to re-configure a few of your gateways to connect to {{% tts %}}, so that you
can start receiving traffic from your end devices.

In order to change a gateway to connect to {{% tts %}}, first you need to
create it using the [Console]({{ ref "/getting-started/console/create-gateway" }})
or the [CLI]({{ ref "/getting-started/cli#create-gateway" }}). Then, you will
need to reconfigure the gateway to connect to {{% tts %}}, and regenerate its
API key (if required).

### Suggested migration process

First, update your applications to work with the V3 data format. If you are
using payload formatters, make sure to set them correctly from the Application
settings page.

Follow the rest of the guide and start by migrating a small number of test end
devices (and gateways, if needed) to {{% tts %}}. Once you are confident that
your devices are working properly, you can start migrating more of your devices
to {{% tts %}}.
