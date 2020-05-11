---
title: "Create Gateway"
description: ""
weight: 3
---

This section contains instructions for creating a Gateway using the command-line interface.

<!--more-->

First, list the available frequency plans:

```bash
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

This creates a gateway `gtw1` with user `admin` as collaborator, frequency plan `EU_863_870`, EUI `00800000A00009EF` and respecting duty-cycle limitations. You can now connect your gateway to {{% tts %}}.

>Note: The CLI returns the created and updated entities by default in JSON. This can be useful in scripts.

### Create Gateway API Key

Some gateways require an API Key with Link Gateway Rights to be able to connect to {{% tts %}}.

Create an API key for the gateway:

```bash
$ ttn-lw-cli gateways api-keys create \
  --name link \
  --gateway-id gtw1 \
  --right-gateway-link
```

The CLI will return an API key such as `NNSXS.VEEBURF3KR77ZR...`. This API key has only link rights and can therefore only be used for linking this gateway. Make sure to copy the key and save it in a safe place. You will not be able to see this key again in the future, and if you lose it, you can create a new one to replace it in the gateway configuration.
