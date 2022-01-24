// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

const DEFAULT_UPLINK_JS_FORMATTER = `function decodeUplink(input) {
  return {
    data: {
      bytes: input.bytes
    },
    warnings: [],
    errors: []
  };
}`
const DEFAULT_DOWNLINK_JS_FORMATTER = `function encodeDownlink(input) {
  return {
    bytes: [],
    fPort: 1,
    warnings: [],
    errors: []
  };
}

function decodeDownlink(input) {
  return {
    data: {
      bytes: input.bytes
    },
    warnings: [],
    errors: []
  }
}`

const REPOSITORY_UPLINK_FORMATTER = `// input = { fPort: 1, bytes: [1, 62] }
function decodeUplink(input) {
  switch (input.fPort) {
  case 1:
    return {
      // Decoded data
      data: {
        direction: directions[input.bytes[0]],
        speed: input.bytes[1]
      }
    }
  default:
    return {
      errors: ["unknown FPort"]
    }
  }
}`

const REPOSITORY_DOWNLINK_FORMATTER = `// input = { data: { led: "green" } }
function encodeDownlink(input) {
  var i = colors.indexOf(input.data.led);
  if (i === -1) {
    return {
      errors: ["invalid LED color"]
    }
  }
  return {
    // LoRaWAN FPort used for the downlink message
    fPort: 2,
    // Encoded bytes
    bytes: [i]
  }
}

// input = { fPort: 2, bytes: [1] }
function decodeDownlink(input) {
  switch (input.fPort) {
  case 2:
    return {
      // Decoded downlink (must be symmetric with encodeDownlink)
      data: {
        led: colors[input.bytes[0]]
      }
    }
  default:
    return {
      errors: ["invalid FPort"]
    }
  }
}`

export const getDefaultJavascriptFormatter = uplink =>
  uplink ? DEFAULT_UPLINK_JS_FORMATTER : DEFAULT_DOWNLINK_JS_FORMATTER

export const getDefaultGrpcServiceFormatter = () => undefined

export const getRepositoryJavascriptFormatter = uplink =>
  uplink ? REPOSITORY_UPLINK_FORMATTER : REPOSITORY_DOWNLINK_FORMATTER
