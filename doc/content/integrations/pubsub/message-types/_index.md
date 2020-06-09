---
title: "Message Types"
description: ""
---

The Application Server can be configured to **publish** messages for any of the following events:
- Downlink ack
- Downlink failed
- Downlink nack
- Downlink queued
- Downlink sent
- Join accept
- Location solved
- Service data
- Uplink message

The Application Server can be configured to **subscribe** to messages to schedule the following events:
- Downlink queue push
- Downlink queue replace

 Enabling event messaging also allows you to manually configure a **Sub topic** for that event. If no Sub topic is specified, events will be published to the configured **Base topic**.

<!--more-->

>Separate Sub topics should be specified for **Downlink queue push** and **Downlink queue replace** in order to use both. Using the Base topic for both simultaneously will cause messages to randomly be scheduled as either **Downlink queue push** or **Downlink queue replace**.

{{< figure src="topics.png" alt="Topics" >}}

## Message Format

JSON messages sent or received by the Application Server are defined in [Data Formats]({{< ref "/integrations/data-formats" >}}).
