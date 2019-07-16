---
title: "Configuring Webhooks"
description: ""
weight: 9
---

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
