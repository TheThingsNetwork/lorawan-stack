---
title: "Rights Management"
description: ""
---

This section contains instructions for managing rights with the CLI.

<!--more-->

## Adding Collaborators

To add collaborators for Gateways, End Devices, or Applications, use the `collaborators add` command. For example, to add a collaborator `user1` to the application `app1` with rights to read and write device info and delete applications, use the command

```bash
$ ttn-lw-cli applications collaborators set --application-id app1 \
  --user-id user1 \
  --right-application-delete \
  --right-application-devices-read \
  --right-application-devices-write
```

> **TIP:** To see the list of possible rights for an entity, use the `--help` flag, e.g `$ ttn-lw-cli applications collaborators set --help`.

## Listing Collaborators

To see which rights a user has on an entity, use the `collaborators list` command. For example, to see collaborators for the gateway `gateway1`, use the command:

```bash
$ ttn-lw-cli gateways collaborators list --gateway-id gateway1
```
