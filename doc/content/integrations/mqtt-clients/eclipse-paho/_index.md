---
title: "Eclipse Paho"
description: ""
weight: 
---

[Eclipse Paho](https://www.eclipse.org/paho/) is an umbrella project on a mission to provide high quality implementations of tools and libraries for M2M communications. It covers MQTT client implementations in several programming languages such as Java, Python, Go, etc.

<!--more-->

Follow this guide to learn how to connect to {{% tts %}} MQTT Server, to receive and to send messages using the Eclipse Paho client.

>This document contains instructions to use [Eclipse Paho MQTT Python client library](https://www.eclipse.org/paho/index.php?page=clients/python/index.php), which implements MQTT v3.1 and v3.1.1 protocol. To compare this library with other Paho project implementations visit [Eclipse Paho Downloads page](https://www.eclipse.org/paho/index.php?page=downloads.php). To find more about the usage of the Python implementation, visit [this page](https://pypi.org/project/paho-mqtt/).

## Requirements

1. Python v2.7 or v3.x installed on your system.

2. [Eclipse Paho MQTT Python client](https://github.com/eclipse/paho.mqtt.python) installed on your system.

## Subscribing to Upstream Traffic

>Note: this section follows the example for subscribing to upstream traffic in [MQTT Server]({{< ref "/integrations/mqtt#subscribing-to-upstream-traffic" >}}) guide.

>To keep things simple, you can use the existing Python scripts from the **examples** folder contained in your installation folder and adjust them according to your setup. 

Enter the **examples** folder and create a new file named `subscribe.py`.

Open the file you created, paste the code below and modify it to match your setup:

```bash 
import context 
import paho.mqtt.subscribe as subscribe

m = subscribe.simple(topics=['#'], hostname="thethings.example.com", port=1883, auth={'username':"app1",'password':"NNSXS.VEEBURF3KR77ZR.."}, msg_count=2)
for a in m:
    print(a.topic)
    print(a.payload)
```

Save the file and run it with:

```bash
$ python subscribe.py
```

Running this script will show the most recent `msg_count` messages published during the last 60 seconds.

To use TLS for security, change the port value to `8883` and pass the `tls` argument to the `simple` function according to its [definition](https://pypi.org/project/paho-mqtt/#id4), where `tls` has to at least contain the path to the CA certificate for your deployment.

## Publishing Downlink Messages

>Note: this section follows the example for publishing downlink messages in [MQTT Server]({{< ref "/integrations/mqtt" >}}) guide. See [Publishing Downlink Messages]({{< ref "/integrations/mqtt#publishing-downlink-traffic" >}}) for a list of available topics.

Create a new file named `publish.py` in the **examples** folder.

Open the file you created and paste the code below:

```bash 
import context
import paho.mqtt.publish as publish

publish.single("v3/{application-id}/devices/{device-id}/down/push", '{"downlinks":[{"f_port": 15,"frm_payload":"vu8=","priority": "NORMAL"}]}', hostname="thethings.example.com", port=1883, {'username':"app1",'password':"NNSXS.VEEBURF3KR77ZR.."})
```

Save the file and run it the terminal with:

```bash
$ python publish.py
```

You will see the scheduled message in the console under the **Live data** tab and your end device will receive the message after a short time.

In case of using TLS, adjust the `port` value and pass the `tls` argument to the `simple` function as described in section above.