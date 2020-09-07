---
title: "InfluxDB Cloud 2.0 Setup"
description: ""
weight: 1
---

Follow the instructions in this section to prepare InfluxDB Cloud 2.0 setup for integrating with {{% tts %}}.

<!--more-->

Log in to your InfluxDB Cloud 2.0 account and select the **Data** on the left hand menu. 

{{< figure src="influxdb-data-dashboard.png" alt="InfluxDB Data dashboard" >}}

On the **Buckets** tab, click the **Create Bucket** button to create a new bucket.

Give your bucket a name, choose how long the data will remain in the database and finish by clicking **Create**.

{{< figure src="creating-a-bucket.png" alt="Creating a bucket" >}}

Next, on the **Tokens** tab, select **Generate** to generate a new **Read/Write Token**.

Enter the **Description** and select the bucket you wish to enable reading and writing for.

{{< figure src="generating-a-token.png" alt="Generating a read/write token" >}}

Go to the **Telegraf** tab and select **Create Configuration**. 

When asked **What do you want to monitor?**, select **System**. 

{{< figure src="monitoring-system.png" alt="Selecting to monitor a system" >}}

Name your configuration, select **Create and Verify** and then **Finish**.

Once the configuration is created, you can simply click on its name in the **Telegraf** tab and download the configuration file. You can further edit this file for usage with the MQTT or Webhook integration.

{{< figure src="telegraf-config.png" alt="Auto-generated Telegraf configuration" >}}
