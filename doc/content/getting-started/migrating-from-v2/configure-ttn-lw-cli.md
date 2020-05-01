---
title: Configure ttn-lw-cli
weight: 16
description: Install and configure CLI tools
---

You will also need the latest version of `ttn-lw-cli`, the CLI for {{% tts %}}.

Get the latest version of `ttn-lw-cli` by following instructions for the [Command-line Interface]({{< ref "getting-started/cli" >}}).



Configure `ttn-lw-cli` for connecting to {{% tts %}}.

```bash
# for Cloud Hosted:
$ ttn-lw-cli use "<tenant-id>.<region>.cloud.thethings.industries"

# or for a private deployment on `thethings.example.com`
# (if you followed the the Getting Started guide):
$ ttn-lw-cli use "thethings.example.com"
```

Login to the {{% tts %}} with the next command, and follow the instructions.

```bash
$ ttn-lw-cli login --callback=false
```

> **NOTE**: The commands above assume that ttn-lw-cli can be found in your PATH. If that
> is not the case, you need to specify the path to `ttn-lw-cli` instead, e.g. `./ttn-lw-cli`.
