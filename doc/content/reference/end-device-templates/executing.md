---
title: "Executing Templates"
description: ""
weight: 4
---

{{< cli-only >}}

Once you created or converted a template, you can execute the template with the `end-device templates execute` command to obtain an end device that can be created.

```bash
# The examples use a template that has been created as follows:
$ ttn-lw-cli end-device template extend \
  --lorawan-version 1.0.3 \
  --lorawan-phy-version 1.0.3-a \
  --frequency-plan-id US_902_928 > example.json
```

You can execute an end device as follows:

```bash
$ cat example.json | ttn-lw-cli end-device template execute
```

<details><summary>Show output</summary>
```json
{
  "ids": {
    "application_ids": {

    }
  },
  "created_at": "0001-01-01T00:00:00Z",
  "updated_at": "0001-01-01T00:00:00Z",
  "lorawan_version": "1.0.3",
  "lorawan_phy_version": "1.0.3-a",
  "frequency_plan_id": "US_902_928"
}
```
</details>

The `end-device template execute` command **does not create** the end device. You can, however, easily pipe the output of `end-device template execute` to create the device. But first, you need to personalize the devices by assigning EUIs since this is a generic device template, see [Assigning EUIs]({{< relref "assigning-euis.md" >}}).

```bash
$ cat example.json \
  | ttn-lw-cli end-device template assign-euis 70b3d57ed0000000 70b3d57ed0000001 \
  | ttn-lw-cli end-device template execute \
  | ttn-lw-cli device create --application-id test-app
```

<details><summary>Show output</summary>
```json
{
  "ids": {
    "device_id": "eui-70b3d57ed0000001",
    "application_ids": {
      "application_id": "test-app"
    },
    "dev_eui": "70B3D57ED0000001",
    "join_eui": "70B3D57ED0000000"
  },
  "created_at": "2019-07-18T12:43:15.937Z",
  "updated_at": "2019-07-18T12:43:16.032962Z",
  "attributes": {
  },
  "network_server_address": "thethings.example.com",
  "application_server_address": "thethings.example.com",
  "join_server_address": "thethings.example.com",
  "lorawan_version": "1.0.3",
  "lorawan_phy_version": "1.0.3-a",
  "frequency_plan_id": "US_902_928",
  "supports_join": true
}
```
</details>
