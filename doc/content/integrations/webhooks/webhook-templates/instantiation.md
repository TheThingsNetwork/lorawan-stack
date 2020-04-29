---
title: "Instantiation"
description: ""
weight: 3
---

The process through which a webhook template becomes a webhook integration is called instantiation. Instantiation is done by the Console after the user has filled in the values of the the template fields. This page describes how the template and the values are combined into the final webhook instance.

<!--more-->

## Instantiation of Header Values

The fields are directly replaced in the values of the headers using the syntax `{field-id}`. Consider the following fragment of a webhook template, describing the available template fields and the headers to be sent to the endpoint:

```yaml
fields:
- id: token
  name: Authentication token
  description: The token used for authentication
  secret: true
  default-value:
headers:
- Authorization: Bearer {token}
```

If the user has filled in the value of `token` with `Zpdc7jWMvYzVTeNQ`, then the resulting webhook will contain a header named `Authorization` with the value `Bearer Zpdc7jWMvYzVTeNQ`.

## Instantiation of URLs and Paths

The fields are replaced inside the URLs and the paths according to the [RFC6570](https://tools.ietf.org/html/rfc6570) format. Consider the following fragment of a webhook template, describing the available template fields and the paths of the endpoint.

```yaml
fields:
- id: username
  name: Username
  description: The username used on the service
  secret: false
  default-value:
- id: create
  name: Create device
  description: If set to true, the device will automatically be created on the first uplink
  secret: false
  default-value: "true"
baseurl: https://www.example.com/lora{/username}
paths:
- uplink-message: /uplink{?create}
```

If the user has filled in the value of `username` with `user1` and the value of `create` with `true`, then the resulting webhook will have its base URL set to `https://www.example.com/lora/user1` and the uplink messages will be sent to `https://www.example.com/lora/user1?create=true` (the uplink messages path will be set to `/uplink?create=true`).
