---
title: "Application Pub-Sub APIs"
description: ""
weight: 4
---

# The `ApplicationPubSubRegistry` service

{{< proto/method service="ApplicationPubSubRegistry" method="GetFormats" >}}

{{< proto/method service="ApplicationPubSubRegistry" method="Set" >}}

{{< proto/method service="ApplicationPubSubRegistry" method="Get" >}}

{{< proto/method service="ApplicationPubSubRegistry" method="List" >}}

{{< proto/method service="ApplicationPubSubRegistry" method="Delete" >}}

# Messages

{{< proto/message message="ApplicationPubSub" >}}

{{< proto/message message="ApplicationPubSub.Message" >}}

{{< proto/message message="ApplicationPubSub.MQTTProvider" >}}

{{< proto/message message="ApplicationPubSub.NATSProvider" >}}

{{< proto/message message="ApplicationPubSubFormats" >}}

{{< proto/message message="ApplicationPubSubIdentifiers" >}}

{{< proto/message message="ApplicationPubSubs" >}}

{{< proto/message message="GetApplicationPubSubRequest" >}}

{{< proto/message message="ListApplicationPubSubsRequest" >}}

{{< proto/message message="SetApplicationPubSubRequest" >}}

# Enums

{{< proto/enum enum="ApplicationPubSub.MQTTProvider.QoS" >}}
