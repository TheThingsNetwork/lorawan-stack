---
title: "Migrating devices from third party LoRaWAN networks"
description: ""
weight: 10
---

This guide documents the process of migrating end devices from third party LoRAWAN networks to {{% tts %}}.

Migrate devices without the need for a rejoin.

## Create a devices.json file

Required fields when importing devices
- device_id
- application_id
- dev_eui
- [join_eui]({{< ref "http://localhost:1313/reference/glossary#joineui" >}}) (also referred to as AppEUI)
- name (optional)
- [lorawan_version]({{< ref "http://localhost:1313/reference/glossary#lorawan-version" >}}) 
- [lorawan_phy_version]({{< ref "reference/glossary#regional-parameters" >}}) (or regional parameters)
- [frequency_plan_id]({{< ref "reference/glossary#frequency-plan" >}})
- supports_join (boolean `true` for OTAA, `false` for ABP devices)
- [app_key]({{< ref "reference/glossary#application-key" >}})
- nwk_key (only for LoRaWAN version 1.1+)
- rx1_delay (optional). Choose between `RX_DELAY_1` or `RX_DELAY_5` for an RX2 delay of 5 seconds (default)
- supports_32_bit_f_cnt (boolean `true` for 32 bit frame counters, `false` for 16 bit counters)
- dev_addr
- app_s_key
- f_nwk_s_int_key (referred to as Network Session Key)
- last_f_cnt_up (optional, frame counter uplinks)
- last_n_f_cnt_down (optional, frame counter downlinks)


Example of `devices.json` file:

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


For more informatino on configurating MAC settings, see [Fine-tuning MAC Settings]({{< ref "getting-started/migrating-to-ttes/configure-mac-settings" >}}).