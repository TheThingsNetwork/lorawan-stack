---
title: "End Devices"
description: ""
weight: 1
---

## List QR Code Formats

To show supported QR code formats for end devices:

```bash
$ ttn-lw-cli end-devices list-qr-formats
```

<details>
<summary>Show output</summary>

```json
{
  "formats": {
    "tr005draft3": {
      "name": "LoRa Alliance TR005 Draft 3",
      "description": "Standard QR code format defined by LoRa Alliance.",
      "field_mask": {
        "paths": [
          "claim_authentication_code.value",
          "ids.dev_eui",
          "ids.join_eui"
        ]
      }
    }
  }
}
```

The formats show the fields of the end device that are used in the QR code.

</details>

## Generate QR Code for Identification

To generate a QR code for identification:

```
$ ttn-lw-cli end-devices generate-qr app1 dev1 --format-id tr005draft3
```

This saves the QR code to the current directory with the device ID as file name, in PNG format with a default size of 300 pixels. Use `--folder` and `--size` to change the save location and image size.

{{< figure src="../qr-identification.png" alt="Device QR Code for Identification" >}}

## Generate QR Code for Claiming

Device claiming is a mechanism to transfer devices securely from one application to another. For example, from a device maker to a device owner, or transferring ownership to new device owner.

When you create a device, you can generate a claim authentication code by specifying the `--with-claim-authentication-code` flag. You can also set a claim authentication code via CLI:

```bash
$ ttn-lw-cli end-devices set app1 dev1 --claim-authentication-code.value=ABCD
```

To generate a QR code for claiming:

```bash
$ ttn-lw-cli end-devices generate-qr app1 dev1 --format-id tr005draft3
```

{{< figure src="../qr-claiming.png" alt="Device QR Code for Claiming" >}}
