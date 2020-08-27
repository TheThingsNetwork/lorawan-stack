---
title: "Javascript"
description: ""
---

Javascript payload formatters allow you to write your own functions to encode or decode messages. Javascript functions are executed using an [JavaScript ECMAScript 5.1](https://www.ecma-international.org/ecma-262/5.1/) engine.

<!--more-->

The payload formatters should be simple and lightweight. Use arithmetic operations and bit shifts to convert binary data to fields. Avoid using non-trivial logic or polyfills. The runtime does not support modules or any input/output other than defined below.

## Decode Uplink Function

The Javascript `decodeUplink()` function is called when a data uplink message is received from a device. This function decodes the binary payload (`frm_payload`) received from the end device to a JSON object (`decoded_payload`) that gets send upstream to the application. The function takes an input object and returns and output object:

```js
function decodeUplink(input) {
  return {
    data: {
      bytes: input.bytes
    }
  };
}
```

The input object has the following structure:

```js
{
  "bytes": [1, 2, 3], // FRMPayload as byte array
  "fPort": 1 // LoRaWAN FPort
}
```

The output object has the following structure:

```js
{
  "data": { ... }, // JSON object
  "warnings": ["warning 1", "warning 2"], // Optional warnings
  "errors": ["error 1", "error 2"] // Optional errors
}
```

> If an error is present in `errors`, the payload is invalid and the message will be dropped. Any warnings in `warnings` are informative.

### Decode Uplink Example: The Things Node

Here is an example `decodeUplink()` function from The Things Node:

```js
function decodeUplink(input) {
  var data = {};
  var events = {
    1: "setup",
    2: "interval",
    3: "motion",
    4: "button"
  };
  data.event = events[input.fPort];
  data.battery = (input.bytes[0] << 8) + input.bytes[1];
  data.light = (input.bytes[2] << 8) + input.bytes[3];
  data.temperature = (((input.bytes[4] & 0x80 ? input.bytes[4] - 0x100 : input.bytes[4]) << 8) + input.bytes[5]) / 100;
  var warnings = [];
  if (data.temperature < -10) {
    warnings.push("it's cold");
  }
  return {
    data: data,
    warnings: warnings
  };
}
```

The following binary input:

```json
{
  "fPort": 4,
  "bytes": [12, 178, 4, 128, 247, 174]
}
```

Yields the data in `decoded_payload` on a data uplink message:

```json
{
  "uplink_message": {
    "f_port": 4,
    "f_cnt": 7825,
    "frm_payload": "DLIEgPeu",
    "decoded_payload": {
      "battery": 3250,
      "event": "button",
      "light": 1152,
      "temperature": -21.3
    },
    "decoded_payload_warnings": ["it's cold"]
  }
}
```

## Encode Downlink Function

The `encodeDownlink()` function is called when a data downlink message is scheduled. This function encodes a JSON object of the downlink message (`decoded_payload`) to binary payload (`frm_payload`) that gets transmitted to the end device. The function takes an input object and returns and output object:

```js
function encodeDownlink(input) {
  return {
    bytes: [1, 2, 3],
    fPort: 1
  }
}
```

The input object has the following structure:

```js
{
  "data": { ... } // JSON object passed by the application as decoded_payload
}
```

The output object has the following structure:

```js
{
  "bytes": [1, 2, 3], // FRMPayload as byte array
  "fPort": 1, // LoRaWAN FPort
  "warnings": ["warning 1", "warning 2"], // Optional warnings
  "errors": ["error 1", "error 2"] // Optional errors
}
```

> If an error is present in `errors`, the payload is invalid and the message will be dropped. Any warnings in `warnings` are informative.

### Encode Downlink Example: The Things Node

Here is an example of an encoder function that uses the `color` field to send a byte on a specific FPort:

```js
function encodeDownlink(input) {
  var colors = ["red", "green", "blue"];
  return {
    bytes: [colors.indexOf(input.data.color)],
    fPort: 4,
  }
}
```

The following data in `decoded_payload` sent by the application:

```json
{
  "data": {
    "color": "blue"
  }
}
```

Yields the following binary output:

```
[ 2 ]
```

And will be sent on FPort 4.

## Decode Downlink Function

The `decodeDownlink()` function is called to decode a data downlink message. This function decodes the binary payload (`frm_payload`) previously encoded with `encodeDownlink()` back to a JSON object (`decoded_payload`). Downlink messages sent upstream as part of events or downlink queue operations are therefore decoded, just like uplink messages.

```js
function decodeDownlink(input) {
  return {
    data: { ... } // JSON object
  }
}
```

The input object has the following structure:

```js
{
  "bytes": [1, 2, 3], // FRMPayload as byte array as returned by encodeDownlink()
  "fPort": 1 // LoRaWAN FPort as returned by encodeDownlink()
}
```

The output object has the following structure:

```js
{
  "data": { ... } // JSON object
}
```

> `decodeDownlink()` must be symmetric with `encodeDownlink()` and should therefore not return any errors.

### Decode Downlink Example: The Things Node

Here is an example of a function that decodes the output of `encodeDownlink()` (see above):

```js
function decodeDownlink(input) {
  switch (input.fPort) {
  case 4:
    return {
      data: {
        color: ["red", "green", "blue"][input.bytes[0]]
      }
    }
  default:
    throw Error("unknown FPort");
  }
}
```

The output of a previous call to `encodeDownlink()`:

```json
{
  "fPort": 4,
  "bytes": [2]
}
```

Yields the following data in the downlink message:

```json
{
  "downlink_message": {
    "f_port": 4,
    "f_cnt": 7825,
    "frm_payload": "Ag==",
    "decoded_payload": {
      "color": "blue"
    }
  }
}
```
