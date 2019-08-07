---
title: "Mapping Templates"
description: ""
weight: 5
---

{{< cli-only >}}

You can use the `end-device templates map` command to map input templates with a mapping file to create new device templates. This allows for combining fields from different device template files.

The matching from input to a mapping template is, in order, by mapping key (`mapping_key`), end device identifiers (`ids.application_id` and `ids.device_id`) and `DevEUI` (`ids.dev_eui`). If you don't specify a mapping key, end device identifiers nor `DevEUI`, the mapping entry always matches. This is useful for mapping many end device templates with a generic template.

Typical use cases are:

1. Assigning identifiers from a mapping file to device templates matching on mapping key.
2. Mapping a device profile (i.e. MAC and PHY versions, frequency plan and class B/C support) from a mapping file to many end device templates.

## Example

This example shows creating a generic device profile that is mapped with provisioning data for secure elements to create device templates.

First, create a mapping file with a device profile:

```
$ ttn-lw-cli end-device template extend \
  --frequency-plan-id EU_863_870 \
  --lorawan-version 1.0.3 \
  --lorawan-phy-version 1.0.3-a \
  --supports-join \
  --supports-class-c \
  --version-ids.brand-id "thethingsproducts" \
  --version-ids.model-id "genericnode" \
  --version-ids.hardware-version "1.0" \
  --version-ids.firmware-version "1.0" > profile.json
```

The mapping file `profile.json` contains the following entries (omitting empty fields).

<details><summary>Show `profile.json`</summary>
```json
{
  "end_device": {
    "version_ids": {
      "brand_id": "thethingsproducts",
      "model_id": "genericnode",
      "hardware_version": "1.0",
      "firmware_version": "1.0"
    },
    "supports_class_c": true,
    "lorawan_version": "1.0.3",
    "lorawan_phy_version": "1.0.3-a",
    "frequency_plan_id": "EU_863_870",
    "supports_join": true
  },
  "field_mask": {
    "paths": [
      "frequency_plan_id",
      "supports_class_c",
      "version_ids.hardware_version",
      "version_ids.model_id",
      "lorawan_phy_version",
      "lorawan_version",
      "supports_join",
      "version_ids.brand_id",
      "version_ids.firmware_version"
    ]
  }
}
```
</details>

Second, convert the provisioning data to a device templates file to `provisioningdata.json`.

>This example uses a **Microchip ATECC608A-MAHTN-T Manifest File**. This file contains provisioning data for The Things Industries Join Server. You can [download the example file](microchip-atecc608a-mahtn-t-example.json).

```bash
$ ttn-lw-cli end-device template from-data microchip-atecc608a-mahtn-t --local-file example.json > provisioningdata.json
```

Third, map the two files to `templates.json`:

```bash
$ cat provisioningdata.json \
  | ttn-lw-cli end-device template map --mapping-local-file profile.json > templates.json
```

This returns the device templates with provisioning data and device profile combined.

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
    "version_ids": {
      "brand_id": "thethingsproducts",
      "model_id": "genericnode",
      "hardware_version": "1.0",
      "firmware_version": "1.0"
    },
    "supports_class_c": true,
    "lorawan_version": "1.0.3",
    "lorawan_phy_version": "1.0.3-a",
    "frequency_plan_id": "EU_863_870",
    "supports_join": true,
    "provisioner_id": "microchip",
    "provisioning_data": {
        "distributor": {
              "organizationName": "Microchip Technology Inc",
              "organizationalUnitName": "Microchip Direct"
            },
        "groupId": "J2D3YNT8Y8WJDC27",
        "manufacturer": {
              "organizationName": "Microchip Technology Inc",
              "organizationalUnitName": "Secure Products Group"
            },
        "model": "ATECC608A",
        ...
        "uniqueId": "0123d34fb176c66f27",
        "version": 1
      }
  },
  "field_mask": {
    "paths": [
      "lorawan_phy_version",
      "supports_join",
      "frequency_plan_id",
      "lorawan_version",
      "version_ids.hardware_version",
      "version_ids.firmware_version",
      "version_ids.model_id",
      "provisioner_id",
      "provisioning_data",
      "version_ids.brand_id",
      "supports_class_c"
    ]
  }
}
{
  "end_device": {
    "ids": {
      "application_ids": {

      }
    },
    "created_at": "0001-01-01T00:00:00Z",
    "updated_at": "0001-01-01T00:00:00Z",
    "version_ids": {
      "brand_id": "thethingsproducts",
      "model_id": "genericnode",
      "hardware_version": "1.0",
      "firmware_version": "1.0"
    },
    "supports_class_c": true,
    "lorawan_version": "1.0.3",
    "lorawan_phy_version": "1.0.3-a",
    "frequency_plan_id": "EU_863_870",
    "supports_join": true,
    "provisioner_id": "microchip",
    "provisioning_data": {
        "distributor": {
              "organizationName": "Microchip Technology Inc",
              "organizationalUnitName": "Microchip Direct"
            },
        "groupId": "J2D3YNT8Y8WJDC27",
        "manufacturer": {
              "organizationName": "Microchip Technology Inc",
              "organizationalUnitName": "Secure Products Group"
            },
        "model": "ATECC608A",
        ...
        "uniqueId": "012350172871677127",
        "version": 1
      }
  },
  "field_mask": {
    "paths": [
      "version_ids.firmware_version",
      "lorawan_phy_version",
      "provisioner_id",
      "provisioning_data",
      "frequency_plan_id",
      "supports_class_c",
      "version_ids.model_id",
      "supports_join",
      "version_ids.brand_id",
      "lorawan_version",
      "version_ids.hardware_version"
    ]
  }
}
```
</details>

Fourth, you can personalize these devices by assigning the `JoinEUI` and `DevEUI` to `devices.json`, see [Assigning EUIs]({{< relref "assigning-euis.md" >}}):

```bash
$ cat templates.json \
  | ttn-lw-cli end-device template assign-euis 70b3d57ed0000000 70b3d57ed0000001 > devices.json
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
    "version_ids": {
      "brand_id": "thethingsproducts",
      "model_id": "genericnode",
      "hardware_version": "1.0",
      "firmware_version": "1.0"
    },
    "supports_class_c": true,
    "lorawan_version": "1.0.3",
    "lorawan_phy_version": "1.0.3-a",
    "frequency_plan_id": "EU_863_870",
    "supports_join": true,
    "provisioner_id": "microchip",
    "provisioning_data": {
        "distributor": {
              "organizationName": "Microchip Technology Inc",
              "organizationalUnitName": "Microchip Direct"
            },
        "groupId": "J2D3YNT8Y8WJDC27",
        "manufacturer": {
              "organizationName": "Microchip Technology Inc",
              "organizationalUnitName": "Secure Products Group"
            },
        "model": "ATECC608A",
        ...
        "uniqueId": "0123d34fb176c66f27",
        "version": 1
      }
  },
  "field_mask": {
    "paths": [
      "provisioner_id",
      "ids.device_id",
      "frequency_plan_id",
      "lorawan_phy_version",
      "provisioning_data",
      "ids.dev_eui",
      "supports_join",
      "version_ids.firmware_version",
      "version_ids.hardware_version",
      "version_ids.brand_id",
      "ids.join_eui",
      "lorawan_version",
      "version_ids.model_id",
      "supports_class_c"
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
    "version_ids": {
      "brand_id": "thethingsproducts",
      "model_id": "genericnode",
      "hardware_version": "1.0",
      "firmware_version": "1.0"
    },
    "supports_class_c": true,
    "lorawan_version": "1.0.3",
    "lorawan_phy_version": "1.0.3-a",
    "frequency_plan_id": "EU_863_870",
    "supports_join": true,
    "provisioner_id": "microchip",
    "provisioning_data": {
        "distributor": {
              "organizationName": "Microchip Technology Inc",
              "organizationalUnitName": "Microchip Direct"
            },
        "groupId": "J2D3YNT8Y8WJDC27",
        "manufacturer": {
              "organizationName": "Microchip Technology Inc",
              "organizationalUnitName": "Secure Products Group"
            },
        "model": "ATECC608A",
        ...
        "uniqueId": "012350172871677127",
        "version": 1
      }
  },
  "field_mask": {
    "paths": [
      "ids.join_eui",
      "ids.dev_eui",
      "version_ids.firmware_version",
      "provisioning_data",
      "frequency_plan_id",
      "provisioner_id",
      "version_ids.model_id",
      "lorawan_version",
      "ids.device_id",
      "version_ids.brand_id",
      "supports_class_c",
      "version_ids.hardware_version",
      "lorawan_phy_version",
      "supports_join"
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
