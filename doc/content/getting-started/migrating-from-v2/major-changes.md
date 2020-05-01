---
title: Major Changes in The Things Stack
weight: 10
---

Before getting started, we will discuss major breaking changes between
{{% ttnv2 %}} and {{% tts %}}, along with some guidelines to make the
migration process easier to manage.

### Application Data

{{% tts %}} uses a different data format for uplink and downlink traffic than {{% ttnv2 %}}.

Example uplink message on {{% ttnv2 %}}:

```js
/* topic name: app1/devices/dev1/up */
{"app_id":"app1","dev_id":"dev1","hardware_serial":"1122334411223344","port":1,"counter":0,"payload_raw":"EQ==","payload_fields":{"led":17},"metadata":{"time":"2020-05-01T00:04:41.258830149Z","latitude":47.984,"longitude":43.123,"altitude":100}}
```

Example downlink message on {{% ttnv2 %}}:

```js
/* topic name: app1/devices/dev1/down */
{"port":10,"confirmed":true,"payload_raw":"EBA="}
```

The data format has changed in {{% tts %}}. It uses a different schema,
different names, and has much richer metadata support. Read more about
that in [Data Formats]({{% ref "/integrations/data-formats" %}}) and in
[Working With Data]({{% ref "/getting-started/working-with-data" %}}).

When migrating to {{% tts %}}, ensure your application can properly handle the new {{% tts %}} data format.

### Payload formats

{{% ttnv2 %}} has support for payload decoders, converters, validators (for uplink) and encoders (for downlink). These can be either CayenneLP or Javascript functions.

{{% tts %}} has support for an uplink payload formatter (similar to the
payload decoder) and a downlink payload formatter (similar to the payload
encoder). These can be set per application, and can even be overridden per end device.

Migrating the {{% ttnv2 %}} payload encoder and decoder to an uplink and downlink payload formatter should be
straightforward, since they have the same format.

### LoRaWAN support

{{% tts %}} requires the LoraWAN MAC Protocol Version and Regional Parameters
(LoraWAN PHY version) to be set per device. These default to `MAC_V1_0_2` and
`MAC_PHY_V1_0_2_REV_B` for devices imported from {{% ttnv2 %}}.

The LoraWAN MAC settings are configurable per device instead of per application.

### MQTT Traffic

You will need to change the MQTT server your application connects to. {{% tts %}} has a new MQTT server address. You will also need to create API keys and update
your MQTT credentials accordingly.

[Read more about using the MQTT Server]({{< ref "/integrations/mqtt" >}}).

{{% tts %}} also supports HTTP webhooks and PubSubs for nats.io or MQTT.

### Storage Integration

{{% tts %}} does not currently support a Storage integration similar to {{% ttnv2 %}}. This feature will be added in a future release.

### Gateway coverage

The Packet Broker enables peering between networks, so traffic received by one
network (e.g. the public community network) but intended for a different
network ({{% tts %}}) can be forwarded to and from that network.

With Packet Broker enabled (which is not discussed here), you can receive
traffic on {{% tts %}} without the need to configure any of your gateways (since
traffic will be automatically forwarded from the public community network).

For private {{% tts %}} deployments with Packet Broker disabled, you will need
to re-configure your gateways to connect to {{% tts %}}, so that you
can start receiving traffic from your end devices.

In order to connect a gateway to {{% tts %}}, follow instructions for [Adding a Gateway in the Console]({{< ref "/getting-started/console/create-gateway" >}})
or [Adding a Gateway using the CLI]({{< ref "/getting-started/cli#create-gateway" >}}). Then, reconfigure the gateway to connect to {{% tts %}}, and regenerate its
API key (if required).

### Suggested migration process

First, update applications to support the {{% tts %}} data format. If you are
using payload formatters, make sure to set them correctly from the Application
settings page.

Follow the rest of the guide and start by migrating a small number of test end
devices (and gateways, if needed) to {{% tts %}}. Once you are confident that
your devices are working properly, migrate the rest
to {{% tts %}}.
