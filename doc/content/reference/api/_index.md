---
title: "API"
description: ""
---

This is the reference for the gRPC and HTTP APIs that {{% tts %}} exposes.

<!--more-->

## Field Masks

{{% tts %}} APIs use field masks to specify a subset of fields that should be returned by a reading request, or to specify fields that should be updated in a writing request. See Google's [Protocol Buffers reference](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#google.protobuf.FieldMask) for more information about field masks.

## HTTP Queries

Additional fields may be specified in HTTP requests by appending them as query string parameters. For example, to request the `name`, `description`, and `locations` of devices in an `EndDeviceRegistry.Get` request, add these fields to the `field_mask` field. To get this data for device `dev1` in application `app1`:

```bash
curl -i -H "Authorization: Bearer NNSXS.XXXXXXXXX" https://thethings.example.com/api/v3/applications/app1/devices/dev1?field_mask=name,description,locations
```

> Fields that are empty or zero are not returned in requests, even if they are specified in a field mask. 

> If you're having trouble with the HTTP API, you can always inspect requests in the Console using your browser's inspector. All of the data displayed in the Console is pulled using HTTP API requests, and this should give you some insight in to how they are formed.
