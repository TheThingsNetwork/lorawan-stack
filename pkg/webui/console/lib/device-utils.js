// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import randomByteString from '@console/lib/random-bytes'

export const ACTIVATION_MODES = Object.freeze({
  OTAA: 'otaa',
  ABP: 'abp',
  MULTICAST: 'multicast',
  NONE: 'none',
})

export const DEVICE_CLASSES = Object.freeze({
  CLASS_A: 'a',
  CLASS_B: 'b',
  CLASS_C: 'c',
})

export const PHY_V1_0 = { value: 'PHY_V1_0', label: 'TS001 Technical Specification 1.0.0' }
export const PHY_V1_0_1 = { value: 'PHY_V1_0_1', label: 'TS001 Technical Specification 1.0.1' }
export const PHY_V1_0_2_REV_A = {
  value: 'PHY_V1_0_2_REV_A',
  label: 'RP001 Regional Parameters 1.0.2',
}
export const PHY_V1_0_2_REV_B = {
  value: 'PHY_V1_0_2_REV_B',
  label: 'RP001 Regional Parameters 1.0.2 revision B',
}
export const PHY_V1_0_3_REV_A = {
  value: 'PHY_V1_0_3_REV_A',
  label: 'RP001 Regional Parameters 1.0.3 revision A',
}
export const PHY_V1_1_REV_A = {
  value: 'PHY_V1_1_REV_A',
  label: 'RP001 Regional Parameters 1.1 revision A',
}
export const PHY_V1_1_REV_B = {
  value: 'PHY_V1_1_REV_B',
  label: 'RP001 Regional Parameters 1.1 revision B',
}
export const RP002_V1_0_0 = { value: 'RP002_V1_0_0', label: 'RP002 Regional Parameters 1.0.0' }
export const RP002_V1_0_1 = { value: 'RP002_V1_0_1', label: 'RP002 Regional Parameters 1.0.1' }
export const RP002_V1_0_2 = { value: 'RP002_V1_0_2', label: 'RP002 Regional Parameters 1.0.2' }
export const RP002_V1_0_3 = { value: 'RP002_V1_0_3', label: 'RP002 Regional Parameters 1.0.3' }

export const LORAWAN_PHY_VERSIONS = Object.freeze([
  PHY_V1_0,
  PHY_V1_0_1,
  PHY_V1_0_2_REV_A,
  PHY_V1_0_2_REV_B,
  PHY_V1_0_3_REV_A,
  PHY_V1_1_REV_A,
  PHY_V1_1_REV_B,
  RP002_V1_0_0,
  RP002_V1_0_1,
  RP002_V1_0_2,
  RP002_V1_0_3,
])

export const LORAWAN_VERSION_PAIRS = {
  100: [PHY_V1_0],
  101: [PHY_V1_0_1],
  102: [PHY_V1_0_2_REV_A, PHY_V1_0_2_REV_B],
  103: [PHY_V1_0_3_REV_A],
  104: [RP002_V1_0_0, RP002_V1_0_1, RP002_V1_0_2, RP002_V1_0_3],
  110: [PHY_V1_1_REV_A, PHY_V1_1_REV_B, RP002_V1_0_0, RP002_V1_0_1, RP002_V1_0_2, RP002_V1_0_3],
  0: LORAWAN_PHY_VERSIONS,
}

export const MAC_V1_0 = { value: 'MAC_V1_0', label: 'LoRaWAN Specification 1.0.0' }
export const MAC_V1_0_1 = { value: 'MAC_V1_0_1', label: 'LoRaWAN Specification 1.0.1' }
export const MAC_V1_0_2 = { value: 'MAC_V1_0_2', label: 'LoRaWAN Specification 1.0.2' }
export const MAC_V1_0_3 = { value: 'MAC_V1_0_3', label: 'LoRaWAN Specification 1.0.3' }
export const MAC_V1_0_4 = { value: 'MAC_V1_0_4', label: 'LoRaWAN Specification 1.0.4' }
export const MAC_V1_1 = { value: 'MAC_V1_1', label: 'LoRaWAN Specification 1.1.0' }

export const LORAWAN_VERSIONS = Object.freeze([
  MAC_V1_0,
  MAC_V1_0_1,
  MAC_V1_0_2,
  MAC_V1_0_3,
  MAC_V1_0_4,
  MAC_V1_1,
])

export const FRAME_WIDTH_COUNT = Object.freeze({
  SUPPORTS_16_BIT: 'supports_16_bit',
  SUPPORTS_32_BIT: 'supports_32_bit',
})

const lwRegexp = /^MAC_V[1-9]_[0-9](_[0-9])?$/
const lwCache = {}

/**
 * Parses string representation of the lorawan mac version to number.
 *
 * @param {string} strMacVersion - Formatted string representation for the
 * lorawan mac version, e.g. 1.1.0.
 * @returns {number} - Number representation of the lorawan mac version. Returns
 * 0 if provided
 * argument is not a valid string representation of the lorawan mac version.
 * @example
 *  const parsedVersion = parseLorawanMacVersion('MAC_V1_0'); // returns 100
 *  const parsedVersion = parseLorawanMacVersion('MAC_V1_0_1'); // returns 110
 *  const parsedVersion = parseLorawanMacVersion(''); // returns 0
 *  const parsedVersion = parseLorawanMacVersion('str'); // returns 0
 */
export const parseLorawanMacVersion = strMacVersion => {
  if (lwCache[strMacVersion]) {
    return lwCache[strMacVersion]
  }

  if (!Boolean(strMacVersion)) {
    return 0
  }

  const match = lwRegexp.exec(strMacVersion)
  if (match === null || match.length === 0) {
    return 0
  }

  const parsed = parseInt(match[0].replace(/\D/g, '').padEnd(3, 0))
  lwCache[strMacVersion] = parsed

  return lwCache[strMacVersion]
}

/**
 * Generates random 16 bytes hex string.
 *
 * @returns {string} - 16 bytes hex string.
 */
export const generate16BytesKey = () => randomByteString(32)

/**
 * Encodes string value of the frame counter width radio field to boolean
 * value necessary for `supports_32_bit_f_ct`.
 *
 * @param {string} value - String representation of `supports_32_bit_f_ct` field,
 * either `supports_16_bit` or `supports_32_bit`.
 * @returns {boolean} - True if value is equal to `supports_32_bit`, false otherwise.
 * @example
 *  const encodedValue = fCntWidthEncode(FRAME_WIDTH_COUNT.SUPPORTS_32_BIT); // returns true
 *  const encodedValue = fCntWidthEncode(FRAME_WIDTH_COUNT.SUPPORTS_16_BIT); // returns false
 */
export const fCntWidthEncode = value => value === FRAME_WIDTH_COUNT.SUPPORTS_32_BIT

/**
 * Decodes boolean value of `supports_32_bit_f_cnt` to string value
 * accepted by the radio field.
 *
 * @param {boolean} value - Value of `supports_32_bit_f_cnt`.
 * @returns {string} - String representation of `supports_32_bit_f_cnt` field.
 * Returns `supports_32_bit` if value is true, `supports_16_bit` if false.
 * @example
 *  const decodedValue = fCntWidthDecode(true); // returns `supports_32_bit`
 *  const decodedValue = fCntWidthDecode(false); // returns `supports_16_bit`
 */
export const fCntWidthDecode = value =>
  value ? FRAME_WIDTH_COUNT.SUPPORTS_32_BIT : FRAME_WIDTH_COUNT.SUPPORTS_16_BIT

const ALL_ZERO_KEY = '0'.repeat(32)

/**
 * @param {string} key - The session key.
 * @returns {boolean} - `true` if the key is non-zero, `false` otherwise.
 */
export const isNonZeroSessionKey = key => key !== ALL_ZERO_KEY

/**
 * @param {string} value - The frequency value.
 * @returns {boolean} - True if the frequency value is valid, false otherwise.
 */
export const dynamicFrequencyTest = value => value === 0 || value >= 100000
