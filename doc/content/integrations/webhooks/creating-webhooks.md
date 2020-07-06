---
title: "Creating Webhooks"
description: ""
weight: -1
---

This section provides instructions for creating a webhook in the console.

<!--more-->

Creating a webhook requires you to have an HTTP(S) endpoint available.

In your application select the **Webhooks** submenu from the **Integrations** side menu. Clicking on the **+ Add Webhook** button will open the Webhook creation screen. Fill in your webhook ID, format and base URL.

{{< figure src="../webhook-creation.png" alt="Webhook creation screen" >}}

The paths are appended to the base URL. So, the Application Server will perform `POST` requests on the endpoint `https://app.example.com/lorahooks/join` for join-accepts and `https://app.example.com/lorahooks/up` for uplink messages. Clicking the **Add Webhook** button will create the Webhook.

>Note: If you don't have an endpoint available for testing, you can test with a free service like [PostBin](https://postb.in).
