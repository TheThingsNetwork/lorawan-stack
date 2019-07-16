---
title: "MQTT API"
description: ""
---

In order to use the MQTT server you need to create a new API key to authenticate:

```bash
$ ttn-lw-cli applications api-keys create \
  --name mqtt-client \
  --application-id app1 \
  --right-application-traffic-read \
  --right-application-traffic-down-write
```

>Note: See `--help` to see more rights that your application may need.

You can now login using an MQTT client with the application ID `app1` as user name and the newly generated API key as password.

There are many MQTT clients available. Great clients are `mosquitto_pub` and `mosquitto_sub`, part of [Mosquitto](https://mosquitto.org).

>Tip: when using `mosquitto_sub`, pass the `-d` flag to see the topics messages get published on. For example:
>
>`$ mosquitto_sub -h localhost -t '#' -u app1 -P 'NNSXS.VEEBURF3KR77ZR..' -d`
