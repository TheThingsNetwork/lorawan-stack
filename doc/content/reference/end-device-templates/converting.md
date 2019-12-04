---
title: "Converting Templates"
description: ""
weight: 3
---

{{< cli-only >}}

You can convert data in various formats to device templates using formats supported by {{% tts %}}. Input formats can be vendor-specific information about devices or device data to migrate from another LoRaWAN server stack.

Start with listing the supported formats:

```bash
$ ttn-lw-cli end-device template list-formats
```

This gives the supported formats. For example:

```json
{
  "formats": {
    "microchip-atecc608a-mahtn-t": {
      "name": "Microchip ATECC608A-MAHTN-T Manifest File",
      "description": "JSON manifest file received through Microchip Purchasing \u0026 Client Services."
    }
  }
}
```

Given input data, you can use the `end-device template from-data` command to get the device template using the specified formatter.

## Example

>This example uses a **Microchip ATECC608A-MAHTN-T Manifest File**. This file contains provisioning data for The Things Industries Join Server. You can [download the example file](microchip-atecc608a-mahtn-t-example.json).

```bash
$ ttn-lw-cli end-device template from-data microchip-atecc608a-mahtn-t --local-file example.json
```

<details><summary>Show output</summary>
```json
{
  "end_device": {
    ...
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
      }
  },
  "field_mask": {
    "paths": [
      "provisioner_id",
      "provisioning_data"
    ]
  },
  "mapping_key": "0123d34fb176c66f27"
}
```
</details>

In this example, only the `provisioner_id` and `provisioning_data` fields are set with the `mapping_key` set to the serial number. Device makers can use the template to assign the `JoinEUI` and `DevEUI`s (see [Assigning EUIs]({{< relref "assigning-euis.md" >}})) as well as other device fields (see [Creating]({{< relref "creating.md" >}}) and [Mapping Templates]({{< relref "mapping.md" >}})).
