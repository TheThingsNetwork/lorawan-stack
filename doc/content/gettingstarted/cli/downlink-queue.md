---
title: "Downlink Queue"
description: ""
weight: 8
---

The stack keeps a queue of downlink messages. Applications can keep pushing downlink messages or replace the queue with a list of downlink messages.

You can see what is in the queue;

```bash
$ ttn-lw-cli end-devices downlink list app1 dev1
```
