---
title: "Command-line Interface"
description: ""
weight: 7
---

The Command-line Interface (CLI) provides a cross-platform interface for managing components through the command-line.

<!--more-->

## Configuration

See the [Configuration Reference]({{< ref "/reference/configuration" >}}) for details on how the CLI can be configured. The different options are detailed in the [CLI Configuration Reference]({{< ref "/reference/configuration/cli.md" >}}).

## Help for Commands

The CLI comes with builtin documentation. This documentation can be printed by adding the `-h` or `--help` flag to any command.

## Command Aliases

Many CLI commands names have aliases, which are different spellings or shorter versions of the same command. For instance, `applications` commands can also be written as `application`, `apps`, `app` or even `a`. This could save you some typing if you frequently use the CLI.

> Tip: For even less typing you could also register an alias for `ttn-lw-cli` itself (`alias tt=ttn-lw-cli`).

## Input and Output Formats

By default the CLI outputs results as one or more JSON objects. These JSON objects are equivalent to the JSON objects that would be returned when using the HTTP API of The Things Stack. Some commands accept one or more JSON objects as input. These JSON objects are equivalent to the JSON objects that can be sent to the HTTP API. This way of handling input and output is used by the [End Device Templates feature]({{< ref "/reference/end-device-templates" >}}).

The `--output-format` flag allows you to specify a [Go template](https://golang.org/pkg/text/template/) that is executed for the result(s) of a command. The example below executes lists applications and outputs their IDs and names:

```bash
$ ttn-lw-cli applications list --name --output-format "{{ .ApplicationID }}: {{ .Name }}"
```

## Login and Logout

The `login` command starts the OAuth authorization flow with the configured OAuth server. By default this makes use of an endpoint on `localhost` for the OAuth callback. If you are not running the CLI on the machine that is `localhost`, you can add the flag `--callback=false` to the `login` command. This will disable the callback endpoint, and instead ask you to copy/paste the authorization code that you will receive from the OAuth server.

It is also possible to use an API key instead of OAuth. To do this, add the flag `--api-key=NNSXS.AAAA.BBBB` to the `login` command (replacing `NNSXS.AAAA.BBBB` with your API key).

Logging out of the CLI is done with the `logout` command, which will automatically invalidate the OAuth access token that was in use.

## Entity Management

In most cases the CLI be used to manage entities, such as applications, end devices or gateways. For each of these entities, the cli has `create`, `get`, `list`, `update` and `delete` commands.

### Creating Entities

When creating an application or gateway, you need to specify a user or organization with the `--user-id` or `--organization-id` flag. This user or organization will be granted all rights on the new application or gateway.

When creating an end device, you need to specify an application in which the end device should be created.

All other flags correspond to the fields of the entity, as can be seen with the `--help` flag and in the [API Reference]({{< ref "/reference/api" >}}).

### Getting or Listing Entities

When getting a single entity or listing entities, you need to provide flags for selecting the fields to fetch from the server. For performance reasons, list commands only allow a subset of fields (the fields stored in the [Identity Server]({{< relref "identity-server.md" >}})).

### Updating Entities

When updating an entity, you provide flags for the fields that need to be updated. All other fields of the entitiy are left unchanged.

### Deleting Entities

Be careful when using the `delete` command, as it won't ask for confirmation.

## Access Control

The CLI can be used to manage access control of entities. Managing collaborators is done with the `collaborators` commands (i.e. `ttn-lw-cli applications collaborators`) and managing API keys with the `api-keys` commands (i.e. `ttn-lw-cli applications api-keys`).

### Collaborator Rights

Collaborator rights can be `set`, `list`ed or `delete`d. When setting collaborator rights, specify a user or organization with the `--user-id` or `--organization-id` flag and the flags for the rights that you want to assign. Use the `--help` or `-h` for more information on the available rights.

### API Keys

API keys can be `create`d, `list`ed, `update`d and `delete`d. Creating an API key is very similar to setting collaborator rights. Be aware that API keys are only shown once when they are created, and will never be shown again, so make sure to copy them to a safe place.

## Uplink Traffic

The CLI can be used to subscribe to uplink traffic. See the [Getting Started guide]({{< ref "/guides/getting-started/cli" >}}) for more details.

## Downlink Queue

The CLI can manage the downlink queue of end devices. See the [Downlink Queue Operations guide]({{< ref "/guides/downlink-queue-ops" >}}) for more details.

## Events

The CLI can be used to subscribe to events. See the [Events guide]({{< ref "/guides/events" >}}) for more details.
