---
title: "Class C"
description: ""
weight: 7
---

In order to send class C downlink messages to a single device, enable class C support for the end device using the following command:

```bash
$ ttn-lw-cli end-devices update app1 dev1 --supports-class-c
```

This will enable the class C downlink scheduling of the device. That's it! New downlink messages are now scheduled as soon as possible.

To disable class C scheduling, set reset with `--supports-class-c=false`.

>Note: you can also pass `--supports-class-c` when creating the device. Class C scheduling will be enable after the first uplink message which confirms the device session.
