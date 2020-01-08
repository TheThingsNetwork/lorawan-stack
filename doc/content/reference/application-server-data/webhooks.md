---
title: "HTTP Webhooks"
description: ""
weight: 20
---

The webhooks feature allows the Application Server to send application related messages to specific HTTP(S) endpoints.

<!--more-->

## Creating a Webhook

Creating a webhook requires you to have an HTTP(S) endpoint available.

In your application select the **Webhooks** submenu from the **Integrations** side menu. Clicking on the **+ Add Webhook** button will open the Webhook creation screen. Fill in your webhook ID, format and base URL.

{{< figure src="../webhook-creation.png" alt="Webhook creation screen" >}}

The paths are appended to the base URL. So, the Application Server will perform `POST` requests on the endpoint `https://app.example.com/lorahooks/join` for join-accepts and `https://app.example.com/lorahooks/up` for uplink messages. Clicking the **Add Webhook** button will create the Webhook.

>Note: If you don't have an endpoint available for testing, use for example [PostBin](https://postb.in).

## Scheduling Downlinks

You can schedule downlink messages using webhooks too. This requires an API key with traffic writing rights, which can be created using the Console. In your application, select the **API Keys** sidemenu and click on the **+ Add API Key** button. You can now fill in the name and the rights of your API key.

{{< figure src="../api-key-creation.png" alt="API key creation screen" >}}

Click on the **Create API Key** button in order to create the API key. This will open the API key information screen.

{{< figure src="../api-key-created.png" alt="API key created" >}}

Make sure to save your API key at this point, since it will no longer be retrievable after you leave the page. You can now pass the API key as bearer token on the `Authorization` header.

The downlink queue operation paths are:

- For push: `/v3/api/as/applications/{application_id}/webhooks/{webhook_id}/devices/{device_id}/down/push`
- For replace: `/v3/api/as/applications/{application_id}/webhooks/{webhook_id}/devices/{device_id}/down/replace`

For example:

```
$ curl https://thethings.example.com/api/v3/as/applications/app1/webhooks/wh1/devices/dev1/down/push \
  -X POST \
  -H 'Authorization: Bearer NNSXS.VEEBURF3KR77ZR..' \
  --data '{"downlinks":[{"frm_payload":"vu8=","f_port":15,"priority":"NORMAL"}]}'
```

Will push a downlink to the end device `dev1` of the application `app1` using the webhook `wh1`.

You can also save the API key in the webhook configuration page using the the **Downlink API Key** field. The Application Server will provide it to your endpoint using the `X-Downlink-APIKey` header and the push and replace operations paths using the `X-Downlink-Push` and `X-Downlink-Replace` headers.
