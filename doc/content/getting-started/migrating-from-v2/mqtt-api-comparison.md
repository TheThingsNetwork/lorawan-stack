---
title: V2 - V3 MQTT API comparison
weight: 70
---

### MQTT Connection Details

| **Category/Type** | **TTI V2** | **TTS V3** |
|---|----|---|
| MQTT Public Address | `{tenant_id}.thethings.industries:1883` | `https://thethings.example.com:1883` (replace with the URL of your deployment) |
| MQTT Public TLS Address | `{tenant_id}.thethings.industries:8883` (we have to use the TLS certificate provided in th e ttn document) | `https://thethings.example.com:1883` |
| Username | `{application_id}` | `{application_id}@{tenant_id}` |
| Password | Application Access Key (with default rights) | Application API Key (with "Write downlink application traffic" and "Read application traffic (uplink and downlink)" rights) |
| Uplink topic (sub) | `{application_id}/devices/{device_id}/up` | `v3/{application id}@{tenant id}/devices/{device id}/up` |
| Downlink scheduling topic (pub) | `{application_id}/devices/{device_id}/downSchedule` the downlink as the first or last item in a downlink queue using "schedule" key value in the payload. | Push downlink - `v3/{application id}@{tenant id}/devices/{device id}/down/pushReplace` downlink queue - `v3/{application id}@{tenant id}/devices/{device id}/down/replace` |


### Uplinks

| **Category/Type** | **TTI V2** | **TTS V3** |
|---|----|---|
| Overall Uplink Format | <details><summary>Show example</summary> {<br>"app_id": "my-app-id",  <br>"dev_id": "my-dev-id",  <br>"hardware_serial": "0102030405060708",<br>"port": 1,  <br>"counter": 2,  <br>"is_retry": false,  <br>"confirmed": false,  <br>"payload_raw": "AQIDBA==",  <br>"payload_fields": {<br>"temperature": 1.0,<br>"luminosity": 0.64<br>},,  <br>"metadata": {<br>"airtime": 46336000,  <br>"time": "1970-01-01T00:00:00.000000000Z",  <br>"frequency": 868.1,  <br>"modulation": "LORA",  <br>"data_rate": "SF7BW125",<br>"coding_rate": "4/5",  <br>"gateways": [<br>{<br>"gtw_id": "ttn-herengracht-ams", <br>"timestamp": 12345,  <br>"time": "1970-01-01T00:00:00.000000000Z", <br>"channel": 2,  <br>"rssi": -25,  <br>"snr": 5,  <br>"rf_chain": 0,  <br>"latitude": 52.1234,  <br>"longitude": 6.1234,  <br>"altitude": 6  <br>},<br>//...more if received by more gateways...<br>],<br>"latitude": 52.2345,  <br>"longitude": 6.2345,  <br>"altitude": 2  <br>}<br>} </details> | <details><summary>Show example</summary> {<br>"end_device_ids" : {<br>"device_id" : "my-dev-id",<br>"application_ids" : {<br>"application_id" : "my-app-id"<br>},<br>"dev_eui" : "0102030405060708",<br>"join_eui" : "0102030405060708",<br>"dev_addr" : "01020304"<br>},<br>"correlation_ids" : [<br>// …<br>],<br>"received_at" : "1970-01-01T00:00:00.000000000Z",<br>"uplink_message" : {<br>"session_key_id" : "AXA50tHUGUucuzS/bCGMNw==",<br>"f_cnt" : 1,<br>"frm_payload" : "AQIDBA==",<br>"decoded_payload" : {<br>"temperature": 1.0,<br>"luminosity": 0.64<br>},<br>"rx_metadata" : [ {<br>"gateway_ids" : {<br>"gateway_id" : "ttn-herengracht-ams",<br>"eui" : "9C5C8E00001A05C4"<br>},<br>"time" : "1970-01-01T00:00:00.000000000Z",<br>"timestamp" : 12345,<br>"rssi" : -25,<br>"channel_rssi" : -25,<br>"snr" : 5,<br>"uplink_token" : "ChIKEAoEZ3R3MRIInFyOAAAaBcQQ6L3Vlgk=",<br>"channel_index" : 2<br>} ],<br>"settings" : {<br>"data_rate" : {<br>"lora" : {<br>"bandwidth" : 125000,<br>"spreading_factor" : 7<br>}<br>},<br>"data_rate_index" : 5,<br>"coding_rate" : "4/5",<br>"frequency" : "868100000",<br>"timestamp" : 12345,<br>"time" : "1970-01-01T00:00:00.000000000Z"<br>},<br>"received_at" : "1970-01-01T00:00:00.000000000Z"<br>}<br>} </details> |
| Gateway fields        | `metadata.gateways` - This object includes gateway details| `uplink_message.rx_metadata` - This object includes gateway details|
|                       | `metadata.gateways.timestamp` - Timestamp when the gateway received the message (microseconds)| `uplink_message.rx_metadata.timestamp` - Gateway concentrator timestamp when the Rx finished (microseconds). |
|                       | `metadata.gateways.time` - Time when the gateway received the message | `uplink_message.rx_metadata.time` - Time when the gateway received the message|
|                       | `metadata.gateways.channel` - Channel where the gateway received the message| As an alternative channel_index is `available.uplink_message.rx_metadata.channel_index` - Index of the gateway channel that received the message. |
|                       | `metadata.gateways.rssi` - Signal strength of the received message| `uplink_message.rx_metadata.rssi` - Signal strength of the received message|
|                       | `metadata.gateways.snr` - Signal to noise ratio of the received message| `uplink_message.rx_metadata.snr` - Signal to noise ratio of the received message|
|                       | `metadata.gateways.rf_chain` - RF chain where the gateway received the message| Not available in documentation |
|                       | `metadata.gateways.latitude` - Latitude of the gateway reported in its status updates | Not available in documentation|
|                       | `metadata.gateways.longitude` - Longitude of the gateway| Not available in documentation |
|                       | `metadata.gateways.altitude` - Altitude of the gateway| Not available in documentation |
|                       | Not available in documentation| `uplink_message.rx_metadata.channel_rssi` - Received signal strength indicator of the channel (dBm)|
|                       | Not available in documentation| `uplink_message.rx_metadata.uplink_token` - Uplink token injected by gateway, Gateway Server or fNS|
|                       |||
| Uplink Metadata       | `metadata` - This object includes metadata| `uplink_message` - this object includes metadata |
|                       | `metadata.airtime` | Not available in documentation |
|                       | `metadata.time` - Time when the server received the message| `uplink_message.received_at` - Server time when the Application Server received the message. |
|                       | `metadata.frequency` - Frequency at which the message was sent| `uplink_message.settings.frequenc`y - Frequency (Hz).|
|                       | `metadata.modulation` - Modulation that was used - LORA or FSK| Not available but as an alternative, modulation can be derived from settings.data_rate |
|                       | `metadata.data_rate` - Data rate that was used - if LORA modulation| `uplink_message.settings.data_rate` - Data rate of LoRa or FSK |
|                       | `metadata.bit_rate` - Bit rate that was used - if FSK modulation| `uplink_message.settings.data_rate.fsk.bit_rate` - Bit rate that was used - if FSK modulation |
|                       | `metadata.coding_rate` - Coding rate that was used| `uplink_message.settings.coding_rate` - LoRa coding rate. |
|                       | `metadata.latitude` - Latitude of the device| Not available in documentation |
|                       | `metadata.longitude` - Longitude of the device| Not available in documentation |
|                       | `metadata.altitude` - Altitude of the device| Not available in documentation |
|                       | `metadata.data_rate_index` - Not available| `uplink_message.data_rate_index` - LoRaWAN data rate index. |
|                       | `metadata.payload_raw` - Base64 encoded payload: [0x01, 0x02, 0x03, 0x04] | `uplink_message.frm_payload` - Frame payload (Base64)|
|                       | `port` - LoRaWAN FPort| `uplink_message.f_port` - LoRaWAN FPort |
|                       | `counter` - LoRaWAN frame counter| `uplink_message.f_cnt` - LoRaWAN frame counter |
|                       | Not available in documentation| `uplink_message.session_key_id` - Join Server issued identifier for the session keys used by this uplink. |
|                       |||
| common fields         | `app_id` | `end_device_ids.application_ids.application_id` |
|                       | `dev_id` | `end_device_ids.device_id` |
|                       | `hardware_serial` - In case of LoRaWAN: the `DevEUI` | `end_device_ids.dev_eui` |
|                       | `confirmed` - is set to true if this message was a confirmed message | Not available in documentation |
|                       | Not available in documentation | `end_device_ids.join_eui` |
|                       | Not available in documentation | `end_device_ids.dev_addr` |
|                       | | |
| Uplink Payload fiedls | `payload_fields` - Object containing the results from the payload functions - left out when empty | `uplink_message.decoded_payload` |
| Correlation IDs       | Not available | `correlation_ids` |

### Downlinks

| **Category/Type** | **TTI V2** | **TTS V3** |
|---|----|---|
| Downlink message overall format | <details><summary>Show example</summary>{<br>"port": 1, <br>"confirmed": false,<br>"payload_raw": "AQIDBA==", <br>"schedule": "replace",<br>}<br><br> or use payload_fields instead <br><br> {<br>"port": 1,<br>"confirmed": false,<br>"payload_fields": {<br>"led": true<br>},<br>"schedule": "replace", // allowed values: "replace" (default), "first", "last"} </details> | <details><summary>Show example</summary> {<br>"downlinks": [{<br>"f_port": 15,<br>"frm_payload": "vu8=",<br>"priority": "HIGH",<br>"confirmed": true,<br>"correlation_ids": ["my-correlation-id"]<br>}]<br>} </details> |
| Individual fields comparison    | `port` | `f_port` |
| | `confirmed` | `confirmed` |
| | `payload_raw` | `frm_payload` |
| | `schedule` | replace and push topics are available as an alternative |
| | Not available | `priority` - if you do not specify a priority, the default priority LOWEST is used. You can specify LOWEST, LOW, BELOW_NORMAL, NORMAL, ABOVE_NORMAL, HIGH and HIGHEST |
| | Not available | `correlation_ids` - You can add multiple custom correlation IDs, for example to reference events or identifiers of your application. | 
| | `payload_fields` | Not available in documentation |

### References

[TTI V2 MQTT API Documentation](https://www.thethingsnetwork.org/docs/applications/mqtt/api.html)

[TTS V3 MQTT API Documentation](https://enterprise.thethingsstack.io/v3.8.1/integrations/mqtt/)
