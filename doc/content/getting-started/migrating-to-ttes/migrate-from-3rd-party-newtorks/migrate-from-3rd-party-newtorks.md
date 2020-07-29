---
title: "Migrating devices from third party LoRaWAN networks"
description: ""
weight: 10
---

This guide documents the process of migrating end devices from third party LoRAWAN networks to {{% tts %}}.

Migrate devices without the need for a rejoin.

## Create a devices.json file

Example

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
            "supports_32_bit_f_cnt":true,
            "resets_f_cnt":false
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


`lorawan_version`
choose between: `1.0.0`, `1.0.1`, `1.0.2`, `1.0.3`, `1.0.4`, `1.1.0`

`lorawan_phy_version`
choose between: `1.0.0`, `1.0.1`, `1.0.2-a`, `1.0.2-b`, `1.0.3-a`, `1.1.0-a`, `1.1.0-b`


frequency_plan_id":"EU_863_870",
supports_join:true




PHY_V1_0_2_REV_B


$ ttn-lw-cli end-devices list-frequency-plans
https://github.com/TheThingsNetwork/lorawan-frequency-plans/blob/master/frequency-plans.yml

frequency_plan_id: choose any of the 
EU_863_870, EU_863_870_TTN, US_902_928_FSB_1, US_902_928_FSB_2, 
AU_915_928_FSB_1
AU_915_928_FSB_2
AU_915_928_FSB_6
CN_470_510_FSB_11
AS_920_923
AS_920_923_LBT
AS_923_925
AS_923_925_LBT
AS_923_925_TTN_AU
KR_920_923_TTN
IN_865_867
RU_864_870_TTN


  "id": "EU_863_870",
  "name": "Europe 863-870 MHz (SF12 for Rx2)",
  "base_frequency": 868
}, {
  "id": "EU_863_870_TTN",
  "base_id": "EU_863_870",
  "name": "Europe 863-870 MHz (SF9 for Rx2 - recommended)",
  "base_frequency": 868
}, {
  "id": "US_902_928_FSB_1",
  "name": "United States 902-928 MHz, FSB 1",
  "base_frequency": 915
}, {
  "id": "US_902_928_FSB_2",
  "name": "United States 902-928 MHz, FSB 2 (used by TTN)",
  "base_frequency": 915
}, {
  "id": "AU_915_928_FSB_1",
  "name": "Australia 915-928 MHz, FSB 1",
  "base_frequency": 915
}, {
  "id": "AU_915_928_FSB_2",
  "name": "Australia 915-928 MHz, FSB 2 (used by TTN)",
  "base_frequency": 915
}, {
  "id": "AU_915_928_FSB_6",
  "name": "Australia 915-928 MHz, FSB 6",
  "base_frequency": 915
}, {
  "id": "CN_470_510_FSB_11",
  "name": "China 470-510 MHz, FSB 11",
  "base_frequency": 470
}, {
  "id": "AS_920_923",
  "name": "Asia 920-923 MHz",
  "base_frequency": 915
}, {
  "id": "AS_920_923_LBT",
  "base_id": "AS_920_923",
  "name": "Asia 920-923 MHz with LBT",
  "base_frequency": 915
}, {
  "id": "AS_923_925",
  "name": "Asia 923-925 MHz",
  "base_frequency": 915
}, {
  "id": "AS_923_925_LBT",
  "base_id": "AS_923_925",
  "name": "Asia 923-925 MHz with LBT",
  "base_frequency": 915
}, {
  "id": "AS_923_925_TTN_AU",
  "base_id": "AS_923_925",
  "name": "Asia 923-925 MHz (used by TTN Australia)",
  "base_frequency": 915
}, {
  "id": "KR_920_923_TTN",
  "name": "South Korea 920-923 MHz",
  "base_frequency": 915
}, {
  "id": "IN_865_867",
  "name": "India 865-867 MHz",
  "base_frequency": 868
}, {
  "id": "RU_864_870_TTN",
  "name": "Russia 864-870 MHz",
  "base_frequency": 868
}, {
  "id": "ISM_2400_3CH_DRAFT2",
  "name": "LoRa 2.4 GHz with 3 channels draft 2",
  "base_frequency": 2450
