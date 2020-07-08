---
title: "Javascript"
description: ""
---

Javascript payload formatters allow you to write your own functions to encode or decode messages. Javascript functions are executed using the [JavaScript ECMAScript 5](https://www.ecma-international.org/ecma-262/5.1/) engine.

<!--more-->

## Decoder Function

The Javascript `Decoder` function is called when an uplink is received from a device. The function has the following format:

```javascript
function Decoder(bytes, f_port) {
  return {
    raw: bytes,
    f_port: f_port
  };
}
```

The `Decoder` function receives the binary message payload as the first parameter (typically called **bytes**) and the port as the second parameter (typically called **f_port**).

The `Decoder` function must return an object, which is appended to the uplink message. The following example is from The Things Node:

```javascript
function Decoder(bytes, port) {
  var decoded = {};
  var events = {
    1: 'setup',
    2: 'interval',
    3: 'motion',
    4: 'button'
  };
  decoded.event = events[port];
  decoded.battery = (bytes[0] << 8) + bytes[1];
  decoded.light = (bytes[2] << 8) + bytes[3];
  decoded.temperature = ((bytes[4] << 8) + bytes[5]) / 100;
  return decoded;
}
```

>Note: the Decoder function should be simple and lightweight. Use arithmetic operations and bit shifts to convert binary data to fields. Avoid using non-trivial logic or polyfills.

The following binary input data:

```
[0x15, 0x4B, 0xA2, 0xD0, 0x5E, 0xDB]
```

produces the following JSON data, appended to the uplink message as a `decoded_payload` field:

```json
{
  "uplink_message": {
    "f_port": 2,
    "f_cnt": 7825,
    "frm_payload": "EiwDzQrL",
    "decoded_payload": {
      "battery": 4652,
      "event": "interval",
      "light": 973,
      "temperature": 27.63
    },
  }
}
```

## Encoder Function

The `Encoder` function is called when a downlink is scheduled. The `Encoder` has the following format

```javascript
function Encoder(payload, f_port) {
  return [];
}
```

The `Encoder` function receives a JSON object as the first parameter (typically called **payload**) and the port as the second parameter (typically called **f_port**).

The function must return a byte encoded array, which will be transmitted as the binary message payload.

For example, if we wish to turn the light on a The Things Node red or green using a `color` field in our payload, we could use the following `Encoder` function:

```javascript
function Encoder(payload, f_port) {
  if(payload.color === "green")
    return [1];
  else if (payload.color === "red")
    return [2];
  else
    return [0]
}
```

>Note: the Encoder function should be simple and lightweight. Use arithmetic operations and bit shifts to convert fields to binary data. Avoid using non-trivial logic or polyfills.

Sending this object:

```javascript
{
  color: 'green'
}
```

will result in the following binary payload being transmitted.

```bash
[ 01 ]
```
