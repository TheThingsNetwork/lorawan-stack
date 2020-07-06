---
title: "Scheduling Downlinks"
description: ""
weight: -1
---

This section provides instructions for creating scheduling downlinks using webhooks.

<!--more-->

You can schedule downlink messages using webhooks. This requires an API key with traffic writing rights, which can be created using the Console. In your application, select the **API Keys** sidemenu and click on the **+ Add API Key** button. You can now fill in the name and the rights of your API key.

{{< figure src="../api-key-creation.png" alt="API key creation screen" >}}

Click on the **Create API Key** button in order to create the API key. This will open the API key information screen.

{{< figure src="../api-key-created.png" alt="API key created" >}}

Make sure to save your API key at this point, since it will no longer be retrievable after you leave the page. You can now pass the API key as bearer token on the `Authorization` header.

The downlink queue operation paths are:

- For push: `/api/v3/as/applications/{application_id}/webhooks/{webhook_id}/devices/{device_id}/down/push`
- For replace: `/api/v3/as/applications/{application_id}/webhooks/{webhook_id}/devices/{device_id}/down/replace`

For example:

```
$ curl https://thethings.example.com/api/v3/as/applications/app1/webhooks/wh1/devices/dev1/down/push \
  -X POST \
  -H 'Authorization: Bearer NNSXS.VEEBURF3KR77ZR..' \
  --data '{"downlinks":[{"frm_payload":"vu8=","f_port":15,"priority":"NORMAL"}]}'
```

Will push a downlink to the end device `dev1` of the application `app1` using the webhook `wh1`.

You can also save the API key in the webhook configuration page using the the **Downlink API Key** field. The Application Server will provide it to your endpoint using the `X-Downlink-Apikey` header and the push and replace operations paths using the `X-Downlink-Push` and `X-Downlink-Replace` headers.
