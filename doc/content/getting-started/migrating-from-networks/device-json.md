---
title: "Create Device.json"
description: ""
weight: 10
---

{{% tts %}} allows you to import devices from other networks using a JSON file describing those devices. Devices imported this way can be migrated without the need for a rejoin.

## Required Fields in `devices.json`

| Field | Type | Description |
|---|---|---|---|
| `ids.device_id` | string | [More info]({{< ref "reference/glossary#device-id" >}}) |
| `ids.application_id` | string | [More info]({{< ref "reference/glossary#application-id" >}}) |
| `ids.dev_eui` | uint64 | [More info]({{< ref "reference/glossary#deveui" >}}) |
| `ids.join_eui` | uint64 | Also referred to as **AppEUI**. [More info]({{< ref "reference/glossary#joineui" >}}) |
| `name` | string | Optional, name of the device |
| `description` | string | Optional, description of the device |
| `lorawan_version` | `defined_only` | e.g.  `MAC_V1_0_2`. [More info]({{< ref "reference/glossary#lorawan-version" >}}) |
| `lorawan_phy_version` | `defined_only` | e.g.  `PHY_V1_0_2_REV_B`. Also referred to as **Regional Parameters Version**. [More info]({{< ref "reference/glossary#regional-parameters" >}}) |
| `frequency_plan_id` | `defined_only` | e.g.  `EU_863_870`. [More info]({{< ref "reference/glossary#frequency-plan" >}}) |
| `supports_join` | boolean | `true` for OTAA, `false` for ABP devices |
| `root_keys.nwk_app.key` | uint128 | Application Key. [More info]({{< ref "reference/glossary#application-key" >}}) |
| `root_keys.nwk_key.key` | uint128 | Network Key. Only for LoRaWAN version 1.1+ |
| `mac_settings.rx1_delay.value` | `RxDelayValue` | Optional. Typical values are `RX_DELAY_1` (1 second) and `RX_DELAY_5` (5 seconds).  [More info]({{< ref "reference/api/end_device#message:MACSettings" >}})|
| `mac_settings.supports_32_bit_f_cnt` | boolean | `true` for 32 bit frame counters, `false` for non-32 bit counters). [More info]({{< ref "reference/api/end_device#message:MACSettings" >}})  |
| `session.dev_addr` | uint32 | Device Address. [More info]({{< ref "reference/glossary#devaddr" >}}) |
| `session.keys.app_s_key.key` | uint128 | Application Session Key. [More info]({{< ref "reference/glossary#application-session-key" >}}) |
| `session.keys.f_nwk_s_int_key.key` | uint128 | Forwarding Network Session Integrity Key, also referred to as **Network Session Key** in LoRaWAN v1.0.x compatibility mode. [More info]({{< ref "reference/api/end_device#message:SessionKeys" >}}) |
| `session.last_f_cnt_up` | int | Optional, frame counter uplink. [More info]({{< ref "reference/api/end_device#message:MACSettings" >}}) |
| `session.last_n_f_cnt_down` | int | Optional, frame counter downlinks. [More info]({{< ref "reference/api/end_device#message:MACSettings" >}}) |

> Note: The dots in the **Field** column imply an embedded object. For example, `root_keys.nwk_app.key` must be set as: 
> ```
> "root_keys": {
>   "nwk_key:": {
>     "key": "<NWK_KEY_HERE>"
>   }
> }, 
> ```

## Example `devices.json`

Below is an example `devices.json` file. The file may contain multiple devices, stored as different JSON objects.

```json
{
  "ids": {
    "device_id": "device-1",
    "application_ids": {
      "application_id": "application-id"
    },
    "dev_eui": "0000000000000000",
    "join_eui": "0000000000000000"
  },
  "name": "name_of_device",
  "description": "description_of_device",
  "lorawan_version":"1.0.2",
  "lorawan_phy_version":"1.0.2-b",
  "frequency_plan_id":"EU_863_870",
  "supports_join":true,
  "root_keys":{
    "app_key":{
      "key":"00000000000000000000000000000000"
    }
  },
  "mac_settings":{
    "rx1_delay":{
      "value":"RX_DELAY_1"
      },
    "supports_32_bit_f_cnt":true
  },
  "session":{
    "dev_addr":"00000000",
    "keys":{
      "app_s_key":{
        "key":"00000000000000000000000000000000"
      },
      "f_nwk_s_int_key":{
        "key":"00000000000000000000000000000000"
      }
    },
    "last_f_cnt_up":0,
    "last_n_f_cnt_down":0
  }
}
{
  "ids": {
    "device_id": "device-2",
    "application_ids": {
      "application_id": "application-id"
    },
    "..."
  }
}
```

For more information on configuring MAC settings, see [Fine-tuning MAC Settings]({{< ref "getting-started/migrating-from-v2/configure-mac-settings" >}}).
