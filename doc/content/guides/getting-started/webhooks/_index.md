---
title: "HTTP Webhooks"
description: ""
weight: 21
---

The webhooks feature allows the Application Server to send application related messages to specific HTTP(S) endpoints.

## Creating webhooks

The Things Stack supports multiple formats to encode messages. To show supported formats, use:

```
$ ttn-lw-cli applications webhooks get-formats
```

The `json` formatter uses the same format as the [MQTT server]({{< relref "../mqtt" >}}).

Creating a webhook requires you to have an HTTP(S) endpoint available, in this example `https://app.example.com/lorahooks`:

```bash
$ ttn-lw-cli applications webhooks set \
  --application-id app1 \
  --webhook-id wh1 \
  --format json \
  --base-url https://app.example.com/lorahooks \
  --join-accept.path /join \
  --uplink-message.path /up
```

This will create a webhook `wh1` for the application `app1` with JSON formatting. The paths are appended to the base URL. So, the Application Server will perform `POST` requests on the endpoint `https://app.example.com/lorahooks/join` for join-accepts and `https://app.example.com/lorahooks/up` for uplink messages.

>Note: You can also specify URL paths for downlink events. See `ttn-lw-cli applications webhooks set --help` for more information.

>Note: If you don't have an endpoint available for testing, use for example [PostBin](https://postb.in).

## Scheduling downlink

You can schedule downlink messages using webhooks too. This requires an API key with traffic writing rights, which can be created as follows:

```bash
$ ttn-lw-cli applications api-keys create \
  --name wh-client \
  --application-id app1 \
  --right-application-traffic-down-write
```

Pass the API key as bearer token on the `Authorization` header.

The path are:

- For push: `/v3/api/as/applications/{application_id}/webhooks/{webhook_id}/devices/{device_id}/down/push`
- For replace: `/v3/api/as/applications/{application_id}/webhooks/{webhook_id}/devices/{device_id}/down/replace`

For example:

```
$ curl https://thethings.example.com/api/v3/as/applications/app1/webhooks/wh1/devices/dev1/down/push \
  -X POST \
  -H 'Authorization: Bearer NNSXS.VEEBURF3KR77ZR..' \
  --data '{"downlinks":[{"frm_payload":"vu8=","f_port":15,"priority":"NORMAL"}]}'
```
