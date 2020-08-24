---
title: "Receive Messages"
description: ""
weight: 2
---

This section shows how to use the Eclipse Paho client to subscribe and listen to messages being published by {{% tts %}} MQTT Server.

<!--more-->

>Note: this section follows the example for subscribing to upstream traffic in [MQTT Server]({{< ref "/integrations/mqtt" >}}) guide.

>Note: to keep things simple, you can use the existing Python scripts from the **examples** folder contained in your installation folder and adjust them according to your setup. 

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

Save the file and run it with 

```bash
python subscribe.py
```
command in the terminal. 

The list containing `msg_count` messages published during the default 60 seconds period will be shown in the terminal as a result of running this script.

In case you want to use TLS for security, change the port value to `8883` and pass the `tls` argument to the `simple` function according to its [definition](https://pypi.org/project/paho-mqtt/#id4), where `tls` has to at least contain the path to the CA certificate for your deployment.