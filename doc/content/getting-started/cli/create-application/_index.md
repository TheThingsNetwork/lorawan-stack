---
title: "Create Application"
description: ""
---

## Create Application

Create the first application:

```bash
$ ttn-lw-cli applications create app1 --user-id admin
```

This creates an application `app1` with the `admin` user as collaborator.

Devices are created within applications.

### Link Application

In order to send uplinks and receive downlinks from your device, you must link the Application Server to the Network Server. In order to do this, create an API key for the Application Server:

```bash
$ ttn-lw-cli applications api-keys create \
  --name link \
  --application-id app1 \
  --right-application-link
```

The CLI will return an API key such as `NNSXS.VEEBURF3KR77ZR...`. This API key has only link rights and can therefore only be used for linking this application. Make sure to copy the key and save it in a safe place. You will not be able to see this key again in the future, and if you lose it, you can create a new one to replace it in the gateway configuration.

You can now link the Application Server to the Network Server:

```bash
$ ttn-lw-cli applications link set app1 --api-key NNSXS.VEEBURF3KR77ZR..
```

Your application is now linked. You can now use the builtin MQTT server and webhooks to receive uplink traffic and send downlink traffic.