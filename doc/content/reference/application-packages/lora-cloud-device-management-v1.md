---
title: "LoRa Cloud Device Management v1"
description: ""
weight: 2
---


The LoRa Cloud Device Management v1 application package communicates the uplinks received from a DM-compatible device to the LoRa Cloud Device Management Service, and schedules the downlinks received from the service back to the device.

More information on the LoRa Cloud Device Management can be found in the [official documentation](https://www.loracloud.com/documentation/device_management?url=overview.html).

## Creating a New Uplink Token

In order to use the LoRa Cloud Device Management application package a new access token must be created in order to allow the Application Server to send the uplinks to the Device Management Service. The new token can be created in the LoRa Cloud Device Management portal, in the section **Token Management**.

{{< figure src="../lora-dms-token-creation.png" alt="Token creation" >}}

After filling in the token name and clicking the **Add New Token** button, the token will be created.

{{< figure src="../lora-dms-token-created.png" alt="Token created" >}}

## Enabling the Package

{{< cli-only >}}

The package can now be enabled using the `associations set` command:

```bash
# Create a JSON formatted file containing the uplink token
$ echo '{ "token": "AQEAdqwV67..." }' > package-data.json
# Create the association
$ ttn-lw-cli applications packages associations set app1 dev1 200 --data-local-file package-data.json
```

This will enable the package on FPort `200` of the device `dev1` of application `app1`. You can now use the LoRa Cloud Device Management in order to manage your device !

<details><summary>Show output</summary>
```json
{
  "ids": {
    "end_device_ids": {
      "device_id": "dev1",
      "application_ids": {
        "application_id": "app1"
      }
    },
    "f_port": 200
  },
  "created_at": "2019-12-18T10:35:15.565807113Z",
  "updated_at": "2019-12-18T22:06:21.693359719Z",
  "package_name": "lora-cloud-device-management-v1",
  "data": {
      "token": "AQEAdqwV67..."
    }
}
```
</details>
