---
title: "Create Device.json"
description: ""
weight: 10
---

{{% tts %}} allows you to import devices from other networks using a JSON file describing those devices. Devices imoported this way can be migrated without the need for a rejoin.

## Required Fields in `devices.json`

- [device_id]({{< ref "reference/glossary#device-id" >}}) 
- [application_id]({{< ref "reference/glossary#application-id" >}}) 
- [dev_eui]({{< ref "reference/glossary#deveui" >}}) 
- [join_eui]({{< ref "reference/glossary#joineui" >}}) (also referred to as AppEUI)
- name (optional)
- [lorawan_version]({{< ref "reference/glossary#lorawan-version" >}}) 
- [lorawan_phy_version]({{< ref "reference/glossary#regional-parameters" >}}) (or regional parameters)
- [frequency_plan_id]({{< ref "reference/glossary#frequency-plan" >}})
- supports_join (boolean `true` for OTAA, `false` for ABP devices)
- [app_key]({{< ref "reference/glossary#application-key" >}})
- nwk_key (only for LoRaWAN version 1.1+)
- [rx1_delay]({{< ref "reference/api/end_device#message:MACSettings" >}}) (optional). Choose between `RX_DELAY_1` or `RX_DELAY_5` for an RX1 delay of 1 or 5 seconds
- [supports_32_bit_f_cnt]({{< ref "reference/api/end_device#message:MACSettings" >}}) (boolean `true` for 32 bit frame counters, `false` for non-32 bit counters)
- [dev_addr]({{< ref "reference/glossary#devaddr" >}})
- [app_s_key]({{< ref "reference/glossary#application-session-key" >}})
- [f_nwk_s_int_key]({{< ref "reference/api/end_device#message:SessionKeys" >}}) (also referred to as Network Session Key)
- [last_f_cnt_up]({{< ref "reference/api/end_device#message:MACSettings" >}}) (optional, frame counter uplinks)
- [last_n_f_cnt_down]({{< ref "reference/api/end_device#message:MACSettings" >}}) (optional, frame counter downlinks)

## Example `devices.json`

Below is an example `devices.json` file. The file may contain multiple devices, stored as an array of the following JSON object.

```json
{
  "ids": {
    "device_id": "device_id",
    "application_ids": {
      "application_id": "application_id"
    },
    "dev_eui": "0000000000000000",
    "join_eui": "0000000000000000"
  },
  "name": "name_of_device",
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
```

For more information on configuring MAC settings, see [Fine-tuning MAC Settings]({{< ref "getting-started/migrating-from-v2/configure-mac-settings" >}}).
