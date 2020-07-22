---
title: "Creating Payload Formatters with the CLI"
description: ""
weight: -1
---

This section explains how to set up Application and device specific payload formatters using the CLI.

<!--more-->

{{< cli-only >}}

## Create an Application Payload Formatter

To create an Application payload formatter, use the following command when linking an Application. If creating a [Javascript payload formatter]({{< relref "javascript" >}}), save your `Encoder` and `Decoder` functions to files and load them using the `formatter-parameter-local-file` parameter:

```bash
$ ttn-lw-cli applications link set app1 \
  --api-key NNSXS.VEEBURF3KR77ZR..
  --default-formatters.down-formatter FORMATTER_JAVASCRIPT \
  --default-formatters.down-formatter-parameter-local-file "encoder.js" \
  --default-formatters.up-formatter FORMATTER_JAVASCRIPT \
  --default-formatters.up-formatter-parameter-local-file "decoder.js"
```

To create a [CayenneLPP]({{< relref "cayenne" >}}) or [Device Repository]({{< relref "device-repo" >}}) Application payload formatter, use the `FORMATTER_CAYENNELPP` or `FORMATTER_DEVICEREPO` constants. No `formatter-parameter-local-file` parameter is needed.

## Create a Device Specific Payload Formatter

It is possible to assign a device specific payload formatter when creating a device using the CLI. Use the following parameters during device creation, and if creating a [Javascript payload formatter]({{< relref "javascript" >}}), save your `Encoder` and `Decoder` functions to files and load them using the `formatter-parameter-local-file` parameter:

```bash
$ ttn-lw-cli end-devices create app1 dev1-with-formatter \
  --dev-eui 0004A30B001C0530 \
  --app-eui 800000000000000C \
  --frequency-plan-id EU_863_870 \
  --root-keys.app-key.key 752BAEC23EAE7964AF27C325F4C23C9A \
  --lorawan-version 1.0.3 \
  --lorawan-phy-version 1.0.3-a \
  --formatters.down-formatter FORMATTER_JAVASCRIPT \
  --formatters.down-formatter-parameter-local-file "encoder.js" \
  --formatters.up-formatter FORMATTER_JAVASCRIPT \
  --formatters.up-formatter-parameter-local-file "decoder.js"
```

To create a [CayenneLPP]({{< relref "cayenne" >}}) or [Device Repository]({{< relref "device-repo" >}}) device payload formatter, use the `FORMATTER_CAYENNELPP` or `FORMATTER_DEVICEREPO` constants. No `formatter-parameter-local-file` parameter is needed.

## Edit a Device Specific Payload Formatter

To change the payload formatter for an existing device, use the `end-devices update` command:

```bash
$ ttn-lw-cli end-devices set app1 dev1-with-formatter \
  --formatters.down-formatter FORMATTER_JAVASCRIPT \
  --formatters.down-formatter-parameter-local-file "encoder.js" \
  --formatters.up-formatter FORMATTER_JAVASCRIPT \
  --formatters.up-formatter-parameter-local-file "decoder.js"
```

To unset the payload formatters, use the `--unset` flag. The command below will unset all device specific payload formatters:

```bash
$ ttn-lw-cli end-devices set app1 dev1-with-formatter \
  --unset "formatters"
```

It is also possible to unset the uplink or downlink formatters separately:

```bash
$ ttn-lw-cli end-devices set app1 dev1-with-formatter \
  --unset "formatters.up-formatter,formatters.up-formatter-parameter"
```
