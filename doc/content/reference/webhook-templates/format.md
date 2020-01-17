---
title: "Format"
description: ""
weight: 2
---

Webhook templates are described using the [YAML](https://yaml.org/) language. Their format is very closely related to that of a normal webhook integration, but with additional fields added.

## Service Description

All of the webhook templates must contain the following fields which describe the service provided by the template to the user.

- `applicationwebhooktemplateidentifiers.templateid`: The unique identifier of the template.
- `name`: The (human readable) name of the service.
- `description`: The description of the service.
- `logourl`: The URL of the logo of the service.
- `infourl`: The URL of the main page of the service.
- `documentationurl`: The URL of the documentation of the service. 

> Note: The difference between `documentationurl` and `infourl` is that `infourl` should lead to the home page of the service (i.e. `https://www.thethingsnetwork.org/`), while `documentationurl` should lead directly to the location of the documentation (i.e. `https://www.thethingsnetwork.org/docs/applications/example/`).

## Template Fields

Templates can contain fields which will be filled by the user on instantiation. The fields are provided as a list named `fields` in the body of the webhook template and contain the following fields:

- `id`: The unique identifier of the field. The ID is only referenced internally and not shown to the user.
- `name`: The (human readable) name of the field.
- `description`: The description of the field.
- `secret`: Controls if the contents of the field should be hidden. To be used in the case of secrets such as passwords, tokens or API keys.
- `defaultvalue`: The value which should be pre filled for the user initially.

For more information on the instantiation process, see [Instantiation]({{< ref "/reference/webhook-templates/instantiation.md" >}}).

## Endpoint

The endpoint of the webhook can be configured using the following fields:

- `format`: The format which the endpoint expects. Currently `json` and `grpc` are supported.
- `headers`: A mapping between the names of the headers and their values. The values can contain template fields.
- `createdownlinkapikey`: Controls if an API Key specific to the service should be created on instantiation.
- `baseurl`: The base URL of the endpoint. Can contain template fields. 
- `uplinkmessage.path`: The path to which uplink messages will be sent. Can contain template fields.
- `joinaccept.path`: The path to which join accept messages will be sent. Can contain template fields.

Status messages received from the Network Server can also be configured using the following fields:

- `downlinkack.path`: The path to which downlink acknowledgements will be sent. Can contain template fields.
- `downlinknack.path`: The path to which downlink not-acknowledged messages will be sent. Can contain template fields.
- `downlinksent.path`: The path to which downlink sent will be sent. Can contain template fields.
- `downlinkfailed.path`: The path to which downlink failures will be sent. Can contain template fields.
- `downlinkqueued.path`: The path to which downlink queued status will be sent. Can contain template fields.
- `locationsolved.path`: The path to which the location of the device will be sent when resolved. Can contain template fields.

> Note: Not all of the messages types must be handled by the service. By setting the field to empty (i.e. `downlinkack:`) the message type will be disabled and the related messages will not be passed to the endpoint.
