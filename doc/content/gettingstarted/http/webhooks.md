---
title: "Using Webhooks"
description: ""
weight: 1
---

## Scheduling Downlink

You can schedule downlink messages using webhooks. The path is `/v3/api/as/applications/{application_id}/webhooks/{webhook_id}/devices/{device_id}/down/push` (or `/replace`). This requires an API key with traffic writing rights, which can be created as follows:

```bash
$ ttn-lw-cli applications api-keys create \
  --name wh-client \
  --application-id app1 \
  --right-application-traffic-down-write
```

Pass the API key as bearer token on the `Authorization` header. For example:

```
$ curl http://localhost:1885/api/v3/as/applications/app1/webhooks/wh1/devices/dev1/down/push \
  -X POST \
  -H 'Authorization: Bearer NNSXS.VEEBURF3KR77ZR..' \
  --data '{"downlinks":[{"frm_payload":"vu8=","f_port":15,"priority":"NORMAL"}]}'
```

### Class C multicast

Multicast messages are downlinks messages which are sent to multiple devices that share the same security context. In the Network Server, this is an ABP session. See [creating a device](#createdev) for learning how to create a multicast device.

Multicast sessions do not allow uplink. Therefore, you need to explicitly specify the gateway(s) to send messages from, using the `class_b_c` field:

```json
{
  "downlinks": [{
    "frm_payload": "vu8=",
    "f_port": 15,
    "priority": "NORMAL",
    "class_b_c": {
      "gateways": [{
        "gateway_ids": {
          "gateway_id": "gtw1"
        }
      }]
    }
  }]
}
```

>Note: if you specify multiple gateways, the Network Server will try the gateways in the order specified. The first gateway with no conflicts and no duty-cycle limitation will send the message.
