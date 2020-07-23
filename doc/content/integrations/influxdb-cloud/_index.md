---
title: "InfluxDB Cloud 2.0"
description: ""
weight: 
---

[InfluxDB Cloud 2.0](https://v2.docs.influxdata.com/v2.0/get-started/) is a serverless real-time monitoring platform, based on data store specifically created for working with time series data. It combines data storage, user interface, visualization, processing, monitoring and alerting into one cohesive system. 

This guide contains instructions on how to configure [Telegraf agent](https://www.influxdata.com/time-series-platform/telegraf/) to listen the messages being published by {{% tts %}} in-built MQTT server and to forward the data into the InfluxDB Cloud 2.0.

<!--more-->

## Requirements

1. A user account on InfluxDB Cloud 2.0.

2. Telegraf agent (version 1.9.2 or higher) [installed](https://portal.influxdata.com/downloads/) on your system.