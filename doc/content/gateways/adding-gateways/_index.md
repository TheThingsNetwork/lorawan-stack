---
title: "Adding Gateways"
description: ""
weight: -1
aliases: [/getting-started/cli/create-gateway, /getting-started/console/create-gateway]
---

This section contains instructions for adding Gateways in {{%tts%}}.

<!--more-->

{{< tabs/container "Console" "CLI" >}}

{{< tabs/tab "Console" >}}

## Adding Gateways using the Console

Go to **Gateways** in the top menu, and click **+ Add Gateway** to reach the gateway registration page. Fill the gateway ID, gateway EUI (if your gateway has an EUI) and frequency plan. The other fields are optional. Click **Create Gateway** to create the gateway.

{{< figure src="gateway-creation.png" alt="Gateway creation" >}}

Your gateway will be created and you will be redirected to the gateway overview page of your newly created gateway.

{{< figure src="gateway-overview.png" alt="Gateway overview" >}}

You can now connect your gateway to {{% tts %}}.

### Create Gateway API Key

Some gateways require an API Key with Link Gateway Rights to be able to connect to {{% tts %}}. 

In order to do this, navigate the **API Keys** menu of your gateway and select **Add API Key**. Enter a name for your key, select the **Link as Gateway to a Gateway Server for traffic exchange, i.e. write uplink and read downlink** right and then press **Create API Key**.

{{< figure src="gateway-api-key-creation.png" alt="Gateway API Key creation" >}}

You will see a screen that shows your newly created API Key. You now can copy it in your clipboard by pressing the clipboard button. After saving the key in a safe place, press **I have copied the key**. You will not be able to see this key again in the future, and if you lose it, you can create a new one to replace it in the gateway configuration.

{{< figure src="gateway-api-key-created.png" alt="Gateway API Key created" >}}

{{< /tabs/tab >}}

{{< tabs/tab "CLI" >}}

## Adding Gateways using the CLI

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

{{< /tabs/tab >}}

{{< /tabs/container >}}

Once a gateway has been added, get started with [Adding Devices]({{< ref "/devices/adding-devices" >}}) and [Integrations]({{< ref "/integrations" >}}) to process and act on data.
