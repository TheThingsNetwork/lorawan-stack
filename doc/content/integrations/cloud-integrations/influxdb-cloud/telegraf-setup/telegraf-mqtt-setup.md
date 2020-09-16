---
title: "Telegraf Setup for MQTT Integration"
description: ""
weight: 1
---

This section contains instructions to configure the Telegraf agent to use [MQTT Consumer](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/mqtt_consumer) plugin for connecting to {{% tts %}} [MQTT Server]({{< ref "/integrations/mqtt" >}}) and to send data to InfluxDB Cloud 2.0.

<!--more-->

The information needed to configure Telegraf can be found on the **MQTT** tab in {{% tts %}} **Integrations** menu.

{{< figure src="../tts-mqtt-info.png" alt="The Things Stack MQTT server info" >}}

Once you have downloaded the Telegraf configuration file as described in [InfluxDB Cloud 2.0 Setup]({{< ref "/integrations/cloud-integrations/influxdb-cloud/influxdb-cloud-setup" >}}), update it by adding the following lines and modifying them according to your MQTT server info:

```bash
[[inputs.mqtt_consumer]]
#
# MQTT broker URLs to be used. The format is scheme://host:port, schema can be tcp, ssl, or ws.
  servers = ["tcp://localhost:1883"]
#
# Topics to subscribe to.
  topics = [
      "#"
      ]
#
# Username and password.
  username = "app-example"
  password = "NNSXS.JNSBLIV34VXYXS7D4ZWV2IKPTGJM3DFRGO6TYDA.OHBQWSVL7Y.........."
#
# Needed only if your payload type is string, since Telegraf does not forward data of this type by default.
  json_string_fields = ["uplink_message_frm_payload"]
#
# Define the message format.
  data_format = "json"
```

Next, you need to copy the previously generated token from the **Tokens** tab and export it to an environmental variable to be used by the InfluxDB **output plugin**, or you can simply pass it directly as a `token` value in the configuration file. You can set the environmental variable by using the following command in your terminal:

```bash
export INFLUX_TOKEN="paste your token here"
```

Run the Telegraf agent in your terminal with the following command:

```bash
telegraf --config /path/to/custom/telegraf.conf
```

## Monitor Your Data

Click the **Explore** tab on the left. Select your bucket in the **FROM** window in the bottom. In the **Filter** window, select **_measurement** on the drop-down menu and tick the **mqtt_consumer** box. 

At this point you will be able to choose which topic and which parameter you want to monitor, and you can start manipulating the incoming data.

{{< figure src="../influxdb-mqtt.png" alt="Monitoring the MQTT data" >}}