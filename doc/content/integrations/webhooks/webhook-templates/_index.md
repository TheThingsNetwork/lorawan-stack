---
title: "Webhook Templates"
description: ""
summary: Webhook templates define a webhook integration that is not created (yet). Templates allows for using common values for many webhooks, such as a common base paths.
weight: 1
---

This is the reference for Webhook Templates

It covers the format of the templates and how the template instantiation process works.

## What is it?

Webhook templates define a webhook integration that is not created (yet). Templates allows for using common values for many webhooks, such as a common base URLs.

## Who is it for?

Webhook templates are primarily targeted at service providers who want to create specialized webhook integrations for the users of {{% tts %}}.

### Typical use cases

1. Create a webhook with a personalized base URL, format and message paths.
2. Provide users with additional information about the webhook itself, using documentation and visual aids.
3. Simplify the process of enabling the integration by removing the manual work of the user.

## How does it work?

Webhook templates can be used to pre fill the common values of a webhook integration such as the base URL, the message paths or the provided headers. They also allow input from the user, in the form of fillable fields, which are then replaced in the template by the Console in order to obtain the concrete webhook. 

See [Template Format]({{< relref "format.md" >}}) for more information on the contents of a webhook template and  [Template Instantiation]({{< relref "instantiation.md" >}}) for more information of the process through which a webhook template, with user input, is converted into a webhook integration.
