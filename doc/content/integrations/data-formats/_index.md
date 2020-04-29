---
title: "Data Formats"
description: ""
---

The streaming data integrations provide multiple formats for the delivery of application uplink messages and downlink queue operations. This section will delve into the common formats used by them.

<!--more-->

## JSON

The JSON message format is described in the sections below.

> Note: Empty fields are omitted from payloads. As such, if a certain field has a value of `""`, `0` or `false` it will not be present in the message.

### Join-accept Messages

The JSON join-accept messages use the following format:

```javascript
{
  "end_device_ids" : {
    "device_id" : "dev1",                    // Device ID
    "application_ids" : {
      "application_id" : "app1"              // Application ID
    },
    "dev_eui" : "0004A30B001C0530",          // DevEUI of the end device
    "join_eui" : "800000000000000C",         // JoinEUI of the end device (also known as AppEUI in LoRaWAN versions below 1.1)
    "dev_addr" : "00BCB929"                  // Device address known by the Network Server
  },
  "correlation_ids" : [ "as:up:01..." ],     // Correlation identifiers of the message
  "received_at" : "2020-02-12T15:15..."      // ISO 8601 UTC timestamp at which the message has been received by the Application Server
  "join_accept" : {
    "session_key_id" : "AXBSH1Pk6Z0G166...", // Join Server issued identifier for the session keys
    "received_at" : "2020-02-17T07:49..."    // ISO 8601 UTC timestamp at which the uplink has been received by the Network Server
  }
}
```

<details><summary>Example join message</summary>

```json
{
  "end_device_ids" : {
    "device_id" : "dev1",
    "application_ids" : {
      "application_id" : "app1"
    },
    "dev_eui" : "0004A30B001C0530",
    "join_eui" : "800000000000000C",
    "dev_addr" : "01497ECC"
  },
  "correlation_ids" : [
    "as:up:01E191YN5F2HMBQCBCVRF731VY",
    "gs:conn:01E191S6ZEWB630NTR45K5QN4Q",
    "gs:uplink:01E191YMZ2S7ZRTEGATVE2S3HH",
    "ns:uplink:01E191YMZ2J64K6FEW5F0WE7TQ",
    "rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E191YMZ2KK3TJYG5C4XZP8JK"
  ],
  "received_at" : "2020-02-17T07:49:09.935284891Z",
  "join_accept" : {
    "session_key_id" : "AXBSH1Pk6Z0G166RlH16CQ==",
    "received_at" : "2020-02-17T07:49:09.736532315Z"
  }
}
```

</details>

### Uplink Messages

The JSON uplink messages use the following format:

```javascript
{
  "end_device_ids" : {
    "device_id" : "dev1",                    // Device ID
    "application_ids" : {
      "application_id" : "app1"              // Application ID
    },
    "dev_eui" : "0004A30B001C0530",          // DevEUI of the end device
    "join_eui" : "800000000000000C",         // JoinEUI of the end device (also known as AppEUI in LoRaWAN versions below 1.1)
    "dev_addr" : "00BCB929"                  // Device address known by the Network Server
  },
  "correlation_ids" : [ "as:up:01...", ... ],// Correlation identifiers of the message
  "received_at" : "2020-02-12T15:15..."      // ISO 8601 UTC timestamp at which the message has been received by the Application Server
  "uplink_message" : {
    "session_key_id": "AXA50...",            // Join Server issued identifier for the session keys used by this uplink
    "f_cnt": 1,                              // Frame counter
    "f_port": 1,                             // Frame port
    "frm_payload": "gkHe",                   // Frame payload (Base64)
    "decoded_payload" : {                    // Decoded payload object, decoded by the device payload formatter
      "temperature": 1.0,
      "luminosity": 0.64
    },
    "rx_metadata": [{                        // A list of metadata for each antenna of each gateway that received this message
      "gateway_ids": {
        "gateway_id": "gtw1",                // Gateway ID
        "eui": "9C5C8E00001A05C4"            // Gateway EUI
      },
      "time": "2020-02-12T15:15:45.787Z",    // ISO 8601 UTC timestamp at which the uplink has been received by the gateway
      "timestamp": 2463457000,               // Timestamp of the gateway concentrator when the message has been received
      "rssi": -35,                           // Received signal strength indicator (dBm)
      "channel_rssi": -35,                   // Received signal strength indicator of the channel (dBm)
      "snr": 5,                              // Signal-to-noise ratio (dB)
      "uplink_token": "ChIKEA...",           // Uplink token injected by gateway, Gateway Server or fNS
      "channel_index": 2                     // Index of the gateway channel that received the message
    }],
    "settings": {                            // Settings for the transmission
      "data_rate": {                         // Data rate settings
        "lora": {                            // LoRa modulation settings
          "bandwidth": 125000,               // Bandwidth (Hz)
          "spreading_factor": 7              // Spreading factor
        }
      },
      "data_rate_index": 5,                  // LoRaWAN data rate index
      "coding_rate": "4/6",                  // LoRa coding rate
      "frequency": "868300000",              // Frequency (Hz)
      "timestamp": 2463457000,               // Timestamp of the gateway concentrator when the message has been received
      "time": "2020-02-12T15:15:45.787Z"     // ISO 8601 UTC timestamp at which the uplink has been received by the gateway
    },
    "received_at": "2020-02-12T15:15..."     // ISO 8601 UTC timestamp at which the uplink has been received by the Network Server
  }
}
```

<details><summary>Example uplink message</summary>

```json
{
  "end_device_ids" : {
    "device_id" : "dev1",
    "application_ids" : {
      "application_id" : "app1"
    },
    "dev_eui" : "0004A30B001C0530",
    "join_eui" : "800000000000000C",
    "dev_addr" : "00BCB929"
  },
  "correlation_ids" : [
    "as:up:01E0WZGT6Y7657CPFPE5WEYDSQ",
    "gs:conn:01E0WDEC6T5T4XXBAX7S1VMFKE",
    "gs:uplink:01E0WZGSZWT07NE5TS2APTV1Z9",
    "ns:uplink:01E0WZGSZXZXZS8RFAWZX0F2FY",
    "rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E0WZGSZXGE1KS577PFBWRJEE"
  ],
  "received_at" : "2020-02-12T15:15:46.014773143Z",
  "uplink_message" : {
    "session_key_id" : "AXA50tHUGUucuzS/bCGMNw==",
    "f_cnt" : 1,
    "frm_payload" : "gkHe",
    "decoded_payload" : {
      "temperature": 1.0,
      "luminosity": 0.64
    },
    "rx_metadata" : [ {
      "gateway_ids" : {
        "gateway_id" : "gtw1",
        "eui" : "9C5C8E00001A05C4"
      },
      "time" : "2020-02-12T15:15:45.787Z",
      "timestamp" : 2463457000,
      "rssi" : -35,
      "channel_rssi" : -35,
      "snr" : 5,
      "uplink_token" : "ChIKEAoEZ3R3MRIInFyOAAAaBcQQ6L3Vlgk=",
      "channel_index" : 2
    } ],
    "settings" : {
      "data_rate" : {
        "lora" : {
          "bandwidth" : 125000,
          "spreading_factor" : 7
        }
      },
      "data_rate_index" : 5,
      "coding_rate" : "4/6",
      "frequency" : "868300000",
      "timestamp" : 2463457000,
      "time" : "2020-02-12T15:15:45.787Z"
    },
    "received_at" : "2020-02-12T15:15:45.789585559Z"
  }
}
```

</details>

### Downlink Events Messages

The JSON downlink `ack`, `nack`, `queued` and `sent` events messages use the following format:

```javascript
{
  "end_device_ids" : {
    "device_id" : "dev1",                    // Device ID
    "application_ids" : {
      "application_id" : "app1"              // Application ID
    },
    "dev_eui" : "0004A30B001C0530",          // DevEUI of the end device
    "join_eui" : "800000000000000C",         // JoinEUI of the end device (also known as AppEUI in LoRaWAN versions below 1.1)
    "dev_addr" : "00BCB929"                  // Device address known by the Network Server
  },
  "correlation_ids" : [ "as:downlink:..." ], // Correlation identifiers of the message
  "received_at" : "2020-02-17T10:32:24...",  // ISO 8601 UTC timestamp at which the message has been received by the Network Server
  "downlink_queued" : {                      // Name of the event (ack, nack, queued or sent)
    "session_key_id" : "AXBSH1Pk6Z0G166...", // Join Server issued identifier for the session keys
    "f_port" : 15,                           // Frame port
    "f_cnt" : 1,                             // Frame counter
    "frm_payload" : "vu8=",                  // Frame payload (Base64)
    "confirmed" : true,                      // If the downlink expects a confirmation from the device or not
    "priority" : "NORMAL",                   // Priority of the message in the downlink queue
    "correlation_ids" : [ "as:downlink..." ] // Correlation identifiers of the message
  }
}
```

<details><summary>Example downlink event message</summary>

```json
{
  "end_device_ids" : {
    "device_id" : "dev1",
    "application_ids" : {
      "application_id" : "app1"
    },
    "dev_eui" : "0004A30B001C0530",
    "join_eui" : "800000000000000C",
    "dev_addr" : "00C30DFA"
  },
  "correlation_ids" : [
    "as:downlink:01E19B99501X84X6CV471TVSZ1",
    "as:up:01E19B9J8ED6HZW5CC41KPGHVD",
    "gs:conn:01E191S6ZEWB630NTR45K5QN4Q",
    "gs:uplink:01E19B9J17SRJWJ4GV3MCJHNWZ",
    "ns:uplink:01E19B9J1CG3QNWS5M8G3DBMPN",
    "rpc:/ttn.lorawan.v3.GsNs/HandleUplink:01E19B9J17K85MJ1MXPHTW0410"
  ],
  "received_at" : "2020-02-17T10:32:24.590413417Z",
  "downlink_ack" : {
    "session_key_id" : "AXBSp8v2brG0jqkri+sIkw==",
    "f_port" : 15,
    "f_cnt" : 1,
    "frm_payload" : "vu8=",
    "confirmed" : true,
    "priority" : "NORMAL",
    "correlation_ids" : [ "as:downlink:01E19B99501X84X6CV471TVSZ1" ]
  }
}
```

</details>

The JSON downlink `failed` event provides both the downlink and the error details, in the following format:

```javascript
{
  "end_device_ids" : {
    "device_id" : "dev1",                    // Device ID
    "application_ids" : {
      "application_id" : "app1"              // Application ID
    }
  },
  "correlation_ids" : [ "as:downlink:..." ], // Correlation identifiers of the message
  "downlink_failed" : {
    "downlink" : {                           // Downlink that which triggered the failure
      "f_port" : 15,                         // Frame port
      "frm_payload" : "YWFhYWFhY...",        // Frame payload (Base64)
      "confirmed" : true,                    // If the downlink expects a confirmation from the device or not
      "priority" : "NORMAL",                 // Priority of the message in the downlink queue
      "correlation_ids" : [ "as:downli..." ] // Correlation identifiers of the message
    },
    "error" : {                              // Error that was encountered while sending the downlink
      "namespace" : "pkg/networkserver",     // Component in which the error occurred
      "name" : "application_downlink_to...", // Error ID
      "message_format" : "application ...",  // Error message
      "correlation_id" : "2e7f786912e94...", // Correlation identifiers of the error
      "code" : 3                             // gRPC error code
    }
  }
}
```

<details><summary>Example downlink failed event message</summary>

```json
{
  "end_device_ids" : {
    "device_id" : "dev1",
    "application_ids" : {
      "application_id" : "app1"
    }
  },
  "correlation_ids" : [ "as:downlink:01E19Z5W3BZBG7CXSJE9C0A3CE" ],
  "downlink_failed" : {
    "downlink" : {
      "f_port" : 15,
      "frm_payload" : "vu8=",
      "confirmed" : true,
      "priority" : "NORMAL",
      "correlation_ids" : [ "as:downlink:01E19Z5W3BZBG7CXSJE9C0A3CE" ]
    },
    "error" : {
      "namespace" : "pkg/networkserver",
      "name" : "application_downlink_too_long",
      "message_format" : "application downlink payload is too long",
      "correlation_id" : "2e7f786912e946e597894e381d0ad7b0",
      "code" : 3
    }
  }
}
```

</details>

### Downlink Messages

The JSON downlink messages use the following format:

```javascript
{
  "end_device_ids" : {
    "device_id" : "dev1",                    // Device ID
    "application_ids" : {
      "application_id" : "app1"              // Application ID
    }
  },
  "downlinks": [{
    "f_port": 15,                            // Frame port
    "frm_payload": "vu8=",                   // Frame payload (Base64)
    "decoded_payload" : {                    // Decoded payload object, to be encoded by the device payload formatter
      "temperature": 1.0,
      "luminosity": 0.64
    },
    "priority": "NORMAL",                    // Priority of the message in the downlink queue
    "confirmed": true,                       // If the downlink expects a confirmation from the device or not
    "correlation_ids" : [ "custom-id" ]      // Correlation identifiers of the message
  }]
}
```

> Note: The end device identifiers need to be supplied only to frontends that do not derive it from other sources (such as the URL for webhooks or topic name for MQTT).

<details><summary>Example downlink message</summary>

```json
{
  "end_device_ids" : {
    "device_id" : "dev1",
    "application_ids" : {
      "application_id" : "app1"
    }
  },
  "downlinks": [{
    "f_port": 15,
    "frm_payload": "vu8=",
    "decoded_payload" : {
      "temperature": 1.0,
      "luminosity": 0.64
    },
    "priority": "NORMAL",
    "confirmed": true,
    "correlation_ids" : [ "custom-id" ]
  }]
}
```

</details>
