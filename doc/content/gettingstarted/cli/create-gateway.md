---
title: "Creating a gateway"
description: ""
weight: 2
---

First, list the available frequency plans:

```
$ ttn-lw-cli gateways list-frequency-plans
```

Then, create the first gateway with the chosen frequency plan:

```bash
$ ttn-lw-cli gateways create gtw1 \
  --user-id admin \
  --frequency-plan-id EU_863_870 \
  --gateway-eui 00800000A00009EF \
  --enforce-duty-cycle
```

This creates a gateway `gtw1` with user `admin` as collaborator, frequency plan `EU_863_870`, EUI `00800000A00009EF` and respecting duty-cycle limitations. You can now connect your gateway to the stack.

>Note: The CLI returns the created and updated entities by default in JSON. This can be useful in scripts.
