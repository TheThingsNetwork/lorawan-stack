---
title: "Telegraf Setup for Webhook Integration"
description: ""
weight: 2
---

This section contains instructions to configure the Telegraf agent to use [HTTP Listener v2](https://github.com/influxdata/telegraf/blob/master/plugins/inputs/http_listener_v2/) plugin for integration with {{% tts %}} via Webhooks and to send data to InfluxDB Cloud 2.0.

<!--more-->

Update the downloaded Telegraf configuration with the following lines:

```bash
[[inputs.http_listener_v2]]
#
# Address and port to host HTTP listener on
  service_address = ":8080"
#
# Path to listen to.
  path = "/telegraf"
#
# HTTP methods to accept.
  methods = ["POST"]
#
# Needed only if your payload type is string, since Telegraf does not forward data of this type by default.
  json_string_fields = ["uplink_message_frm_payload"]
#
# Define the message format.
  data_format = "json"
```

Copy the generated token from the **Tokens** tab and export it to an environmental variable by using the following command in your terminal:

```bash
export INFLUX_TOKEN="paste your token here"
```

Run the Telegraf agent with this configuration file.

In {{% tts %}} Console, [create a new webhook]({{< ref "/integrations/webhooks/creating-webhooks" >}}) with JSON **Webhook format**, set the **Base URL** to `http://localhost:8080/telegraf` and tick the box next to the message types you want to enable this webhook for.

{{< figure src="../tts-webhook-info.png" alt="Creating webhook on The Things Stack" >}}

>Note: keep in mind that Telegraf agent can be hosted in a remote environment as well. In that case, you need to adjust the **Base URL** according to your setup.

Click the **Explore** tab on the left in InfluxDB Cloud 2.0. Select your bucket in the **FROM** window in the bottom. In the **Filter** window, select **_measurement** on the drop-down menu and tick the **http_listener_v2** box. 

In another **Filter** window, you can select the **uplink_message_decoded_payload** and click the **Submit** field on the right to see the incoming data.
