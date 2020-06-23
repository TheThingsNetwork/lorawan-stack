---
title: "Webhook Path Variables"
description: ""
weight: -1
---

Webhook path variables allow you to substitute device and application specific variables in webhook paths. This section provides instructions for using webhook path variables.

<!--more-->

Webhook path variables allow you to use the following variables in webhook paths:

- `appID`
- `appEUI`
- `joinEUI`
- `devID`
- `devEUI`
- `devAddr`

Path variables can be inserted in the **Base URL** webhook field or the **Path** field for a particular type of message.

For example, if the **Base URL** is `https://app.example.com/lorahooks{/appID}` and the **Path** is `/up{/devID}` an uplink from the device `dev1` of application `app1` will be posted at `https://app.example.com/lorahooks/app1/up/dev1`.

See [IETF RFC65700](https://tools.ietf.org/html/rfc6570) for more documentation about URL path variables. {{% tts %}} supports all forms of path variable substitution.
