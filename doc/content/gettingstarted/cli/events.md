---
title: "Receiving events"
description: ""
weight: 6
---

The stack generates lots of events that allow you to get insight in what is going on. You can subscribe to application, gateway, end device events, as well as to user, organization and OAuth client events.

To follow your gateway `gtw1` and application `app1` events at the same time:

```bash
$ ttn-lw-cli events subscribe --gateway-id gtw1 --application-id app1
```
