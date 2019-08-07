---
title: "Assigning EUIs"
description: ""
weight: 6
---

{{< cli-only >}}

You can use the `end-device template assign-euis` command to assign the `JoinEUI` and `DevEUI` to device templates. The `DevEUI` is assigned from the given start number. This personalizes the device templates.

You can use the command in two different ways:

1. Pipe a number of device templates to assign EUIs to, for example the result of [Mapping Templates]({{< relref "mapping.md" >}}) or [Converting Templates]({{< relref "converting.md" >}})
2. Pipe a generic device template, for example the result of [Creating Templates]({{< relref "creating.md" >}}), and pass the `--count` flag to generate a number of personalized device templates

Once the device templates are personalized with EUIs, you can set additional fields (see [Creating Templates]({{< relref "creating.md" >}})) or execute templates (see [Executing Templates]({{< relref "executing.md" >}})).

>Note: to make your device compliant with most LoRaWAN networks, use the `JoinEUI` of the LoRaWAN Join Server where the devices will be registered and use a `DevEUI` from an IEEE issued MAC address block that is assigned to your company. If you intend to operate your own LoRaWAN Join Server, use a `JoinEUI` from your IEEE assigned MAC address block as LoRa Alliance requires that for DNS lookup. [Learn how to apply for a IEEE issued MAC address block](https://standards.ieee.org/products-services/index.html).

## Example

This example illustrates creating a device profile template, assigning 5 `DevEUI`s and creating them in your The Things Stack application.

First, create a mapping file with a device profile in `profile.json`:

```bash
$ ttn-lw-cli end-device template extend \
  --frequency-plan-id EU_863_870 \
  --lorawan-version 1.0.3 \
  --lorawan-phy-version 1.0.3-a \
  --supports-join > profile.json
```

Second, assign the EUIs. The first argument is the `JoinEUI`, the second argument is the first `DevEUI`.

```bash
$ cat profile.json \
  | ttn-lw-cli end-device template assign-euis 70b3d57ed0000000 70b3d57ed0000001 --count 5 > devices.json
```

<details><summary>Show output</summary>
```json
{
  "end_device": {
    "ids": {
      "device_id": "eui-70b3d57ed0000001",
      "application_ids": {

      },
      "dev_eui": "70B3D57ED0000001",
      "join_eui": "70B3D57ED0000000"
    },
    "created_at": "0001-01-01T00:00:00Z",
    "updated_at": "0001-01-01T00:00:00Z",
    "attributes": {
    },
    "lorawan_version": "1.0.3",
    "lorawan_phy_version": "1.0.3-a",
    "frequency_plan_id": "EU_863_870",
    "supports_join": true
  },
  "field_mask": {
    "paths": [
      "lorawan_version",
      "ids.device_id",
      "ids.join_eui",
      "ids.dev_eui",
      "supports_join",
      "frequency_plan_id",
      "lorawan_phy_version"
    ]
  }
}
{
  "end_device": {
    "ids": {
      "device_id": "eui-70b3d57ed0000002",
      "application_ids": {

      },
      "dev_eui": "70B3D57ED0000002",
      "join_eui": "70B3D57ED0000000"
    },
    "created_at": "0001-01-01T00:00:00Z",
    "updated_at": "0001-01-01T00:00:00Z",
    "attributes": {
    },
    "lorawan_version": "1.0.3",
    "lorawan_phy_version": "1.0.3-a",
    "frequency_plan_id": "EU_863_870",
    "supports_join": true
  },
  "field_mask": {
    "paths": [
      "ids.dev_eui",
      "supports_join",
      "frequency_plan_id",
      "lorawan_phy_version",
      "lorawan_version",
      "ids.device_id",
      "ids.join_eui"
    ]
  }
}
{
  "end_device": {
    "ids": {
      "device_id": "eui-70b3d57ed0000003",
      "application_ids": {

      },
      "dev_eui": "70B3D57ED0000003",
      "join_eui": "70B3D57ED0000000"
    },
    "created_at": "0001-01-01T00:00:00Z",
    "updated_at": "0001-01-01T00:00:00Z",
    "attributes": {
    },
    "lorawan_version": "1.0.3",
    "lorawan_phy_version": "1.0.3-a",
    "frequency_plan_id": "EU_863_870",
    "supports_join": true
  },
  "field_mask": {
    "paths": [
      "supports_join",
      "frequency_plan_id",
      "lorawan_phy_version",
      "lorawan_version",
      "ids.device_id",
      "ids.join_eui",
      "ids.dev_eui"
    ]
  }
}
{
  "end_device": {
    "ids": {
      "device_id": "eui-70b3d57ed0000004",
      "application_ids": {

      },
      "dev_eui": "70B3D57ED0000004",
      "join_eui": "70B3D57ED0000000"
    },
    "created_at": "0001-01-01T00:00:00Z",
    "updated_at": "0001-01-01T00:00:00Z",
    "attributes": {
    },
    "lorawan_version": "1.0.3",
    "lorawan_phy_version": "1.0.3-a",
    "frequency_plan_id": "EU_863_870",
    "supports_join": true
  },
  "field_mask": {
    "paths": [
      "ids.device_id",
      "ids.join_eui",
      "ids.dev_eui",
      "supports_join",
      "frequency_plan_id",
      "lorawan_phy_version",
      "lorawan_version"
    ]
  }
}
{
  "end_device": {
    "ids": {
      "device_id": "eui-70b3d57ed0000005",
      "application_ids": {

      },
      "dev_eui": "70B3D57ED0000005",
      "join_eui": "70B3D57ED0000000"
    },
    "created_at": "0001-01-01T00:00:00Z",
    "updated_at": "0001-01-01T00:00:00Z",
    "attributes": {
    },
    "lorawan_version": "1.0.3",
    "lorawan_phy_version": "1.0.3-a",
    "frequency_plan_id": "EU_863_870",
    "supports_join": true
  },
  "field_mask": {
    "paths": [
      "supports_join",
      "frequency_plan_id",
      "lorawan_phy_version",
      "lorawan_version",
      "ids.device_id",
      "ids.join_eui",
      "ids.dev_eui"
    ]
  }
}
```
</details>

Finally, you can create these devices in your The Things Stack application `test-app`, see [Executing Templates]({{< relref "executing.md" >}}).

```bash
$ cat devices.json \
  | ttn-lw-cli end-device template execute \
  | ttn-lw-cli device create --application-id test-app
```
