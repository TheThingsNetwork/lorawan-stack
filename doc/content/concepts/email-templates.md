---
title: "Email Templates"
description: ""
weight: 15
summary: Email templates define the contents of the emails that The Things Stack sends to its users. They allow network operators to override the default contents of the emails with their own custom contents.
---

## What is it?

Email templates define the contents of the emails that {{% tts %}} sends to its users. They allow network operators to override the default contents of the emails with their own custom contents.

## Who is it for?

Email templates are targeted at network operators that would like to customize the emails that {{% tts %}} sends to its users.

### Typical use cases

1. Adding additional styling to the default text-only emails that {{% tts %}} sends.
2. Translating the emails {{% tts %}} sends to its users (applies deployment-wide).

## How does it work?

Email templates override the default emails that {{% tts %}} sends by providing custom template files written in Go's [html/template](https://golang.org/pkg/html/template/) format. The templates are retrieved when {{% tts %}} sends the first email of a certain type, and then are cached until it is restarted. See [Available Templates]({{< ref "/reference/email-templates/available.md" >}}) and [Overriding Default Templates]({{< ref "/reference/email-templates/overriding.md" >}}).
