---
title: "Email Templates"
description: ""
weight: 15
summary: Email templates define the contents of the emails that The Things Stack sends to its users. They allow network operators to override the default contents of the emails with their own custom contents.
---

## What is it?

Email templates define the contents of the emails that The Things Stack sends to its users. They allow network operators to override the default contents of the emails with their own custom contents.

## Who is it for?

Email templates are targeted at network operators that would like to customize the emails that The Things Stack sends to its users.

### Typical use cases

1. Adding additional styling to the default text-only emails that The Things Stack sends.
2. Translating the emails The Things Stack sends to its users (applies deployment-wide).
 
## How does it work?

Email templates override the default emails that The Things Stack sends by providing custom template files written in Go's [html/template](https://golang.org/pkg/html/template/) format. The templates are retrieved when The Things Stack sends the first email of a certain type, and then are cached until it is restarted. See [Available Templates]({{< ref "/reference/email-templates/available.md" >}}) and [Overriding Default Templates]({{< ref "/reference/email-templates/overriding.md" >}}).
