---
title: "InfluxDB Cloud 2.0"
description: ""
weight: 
---

[InfluxDB Cloud 2.0](https://v2.docs.influxdata.com/v2.0/get-started/) is a serverless real-time monitoring platform specifically created for working with time series data. It combines data storage, user interface, visualization, processing, monitoring and alerting into one cohesive system. 

Besides being able to send data to InfluxDB Cloud 2.0, [Telegraf agent](https://www.influxdata.com/time-series-platform/telegraf/) can also be configured to subscribe to messages published by {{% tts %}} [MQTT server]({{< ref "/integrations/mqtt" >}}) or to listen to messages sent by {{% tts %}} Application Server via [HTTP Webhooks]({{< ref "/integrations/webhooks" >}}). This guide contains the instructions for both of these implementations.

<!--more-->

## Requirements

1. A user account on InfluxDB Cloud 2.0.

2. Telegraf agent (version 1.9.2 or higher) [installed](https://portal.influxdata.com/downloads/) on your system.
