---
title: "Creating Templates"
description: ""
weight: 2
---

{{< cli-only >}}

You can create a device template from an existing device or extend an existing device template. You can also create a new template from scratch.

## Create from existing device

You can use the `end-device template create` command to create a template from an existing device.

>Note: By default, `end-device template create` strips the device's application ID, device ID, `JoinEUI`, `DevEUI` and server addresses to create a generic template.
>
>You can include the end device identifiers by passing the concerning flags: `--application-id`, `--device-id`, `--join-eui` and `--dev-eui`.

Pipe the output from getting a device to create a template, for example:

```bash
$ ttn-lw-cli device get test-app test-dev \
  --lorawan-version \
  --lorawan-phy-version \
  | ttn-lw-cli end-device template create > template.json
```

<details><summary>Show output</summary>
```json
{
  "end_device": {
    "ids": {
      "application_ids": {

      }
    },
    "created_at": "0001-01-01T00:00:00Z",
    "updated_at": "0001-01-01T00:00:00Z",
    "lorawan_version": "1.0.3",
    "lorawan_phy_version": "1.0.3-a"
  },
  "field_mask": {
    "paths": [
      "lorawan_version",
      "lorawan_phy_version"
    ]
  }
}
```
</details>

## Extend existing template

Use the `end-device template extend` command to extend a template:

```bash
$ cat template.json \
  | ttn-lw-cli end-device template extend \
  --frequency-plan-id EU_863_870
```

<details><summary>Show output</summary>
```json
{
  "end_device": {
    "ids": {
      "application_ids": {

      }
    },
    "created_at": "0001-01-01T00:00:00Z",
    "updated_at": "0001-01-01T00:00:00Z",
    "attributes": {
    },
    "lorawan_version": "1.0.3",
    "lorawan_phy_version": "1.0.3-a",
    "frequency_plan_id": "EU_863_870"
  },
  "field_mask": {
    "paths": [
      "lorawan_phy_version",
      "frequency_plan_id",
      "lorawan_version"
    ]
  }
}
```
</details>

See `$ ttn-lw-cli end-device template extend --help` for all the fields that can be set.

## Create from scratch

The `end-device template extend` can also be used to create a new template from scratch by simply not piping an existing device as input.

For example, create a new template from scratch:

```bash
$ ttn-lw-cli end-device template extend \
  --lorawan-version 1.0.3 \
  --lorawan-phy-version 1.0.3-a \
  --frequency-plan-id EU_863_870
```

<details><summary>Show output</summary>
```json
{
  "end_device": {
    "ids": {
      "application_ids": {

      }
    },
    "created_at": "0001-01-01T00:00:00Z",
    "updated_at": "0001-01-01T00:00:00Z",
    "attributes": {
    },
    "lorawan_version": "1.0.3",
    "lorawan_phy_version": "1.0.3-a",
    "frequency_plan_id": "EU_863_870"
  },
  "field_mask": {
    "paths": [
      "frequency_plan_id",
      "lorawan_phy_version",
      "lorawan_version"
    ]
  }
}
```
</details>
