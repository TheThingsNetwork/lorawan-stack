---
title: "Send Messages"
description: ""
weight: 3
---

This section shows how to use the Eclipse Paho client to schedule a downlink message to be sent to your end device.

<!--more-->

>Note: this section follows the example for publishing downlink traffic in [MQTT Server]({{< ref "/integrations/mqtt" >}}) guide.

Create a new file named `publish.py` in the **examples** folder.

Open the file you created and paste the code below:

```bash 
import context
import paho.mqtt.publish as publish

publish.single("v3/{application-id}/devices/{device-id}/down/push", '{"downlinks":[{"f_port": 15,"frm_payload":"vu8=","priority": "NORMAL"}]}', hostname="thethings.example.com", port=1883, {'username':"app1",'password':"NNSXS.VEEBURF3KR77ZR.."})
```

Save the file and run it with `python publish.py` in the terminal. 

You will see the downlink message being scheduled in {{% tts %}} Console under **Live data** tab and your end device will receive the message in a short time.

In case of using TLS, adjust the `port` value and pass the `tls` argument to the `simple` function as described in the [Receive Messages]({{< ref "/integrations/mqtt-clients/eclipse-paho/receive" >}}) section.