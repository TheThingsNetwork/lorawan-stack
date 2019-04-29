---
title: "Integrate Things"
description: ""
weight: 7
---

## <a name="webhooks">Using webhooks</a>

The webhooks feature allows the Application Server to send application related messages to specific HTTP(S) endpoints.

To show supported formats, use:

```
$ ttn-lw-cli applications webhooks get-formats
```

The `json` formatter uses the same format as the MQTT server described above.

Creating a webhook requires you to have an HTTP(S) endpoint available.

```bash
$ ttn-lw-cli applications webhooks set \
  --application-id app1 \
  --webhook-id wh1 \
  --format json \
  --base-url https://example.com/lorahooks \
  --join-accept.path /join \
  --uplink-message.path /up
```

This will create a webhook `wh1` for the application `app1` with JSON formatting. The paths are appended to the base URL. So, the Application Server will perform `POST` requests on the endpoint `https://example.com/lorahooks/join` for join-accepts and `https://example.com/lorahooks/up` for uplink messages.

>Note: You can also specify URL paths for downlink events, just like MQTT. See `ttn-lw-cli applications webhooks set --help` for more information.

You can also send downlink messages using webhooks. The path is `/v3/api/as/applications/{application_id}/webhooks/{webhook_id}/devices/{device_id}/down/push` (or `/replace`). Pass the API key as
bearer token on the `Authorization` header. For example:

```
$ curl http://localhost:1885/api/v3/as/applications/app1/webhooks/wh1/devices/dev1/down/push \
  -X POST \
  -H 'Authorization: Bearer NNSXS.VEEBURF3KR77ZR..' \
  --data '{"downlinks":[{"frm_payload":"vu8=","f_port":15,"priority":"NORMAL"}]}'
```

## Congratulations

You have now set up The Things Network Stack V3! 🎉

### Go further

* [Events](../../concepts/events)
