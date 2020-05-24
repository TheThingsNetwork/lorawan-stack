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

export const LORAWAN_VERSIONS = Object.freeze([
  { value: '1.0.0', label: 'MAC V1.0' },
  { value: '1.0.1', label: 'MAC V1.0.1' },
  { value: '1.0.2', label: 'MAC V1.0.2' },
  { value: '1.0.3', label: 'MAC V1.0.3' },
  { value: '1.0.4', label: 'MAC V1.0.4' },
  { value: '1.1.0', label: 'MAC V1.1' },
])

export const LORAWAN_PHY_VERSIONS = Object.freeze([
  { value: '1.0.0', label: 'PHY V1.0' },
  { value: '1.0.1', label: 'PHY V1.0.1' },
  { value: '1.0.2-a', label: 'PHY V1.0.2 REV A' },
  { value: '1.0.2-b', label: 'PHY V1.0.2 REV B' },
  { value: '1.0.3-a', label: 'PHY V1.0.3 REV A' },
  { value: '1.1.0-a', label: 'PHY V1.1 REV A' },
  { value: '1.1.0-b', label: 'PHY V1.1 REV B' },
])

const lwRegexp = /^[1-9].[0-9].[0-9]$/
const lwCache = {}

/**
 * Parses string representation of the lorawan mac version to number.
 *
 * @param {string} strMacVersion - Formatted string representation fot the
 * lorawan mac version, e.g. 1.1.0.
 * @returns {number} - Number representation of the lorawan mac version. Returns
 * 0 if provided
 * argument is not a valid string representation of the lorawan mac version.
 * @example
 *  const parsedVersion = parseLorawanMacVersion('1.0.0'); // returns 100
 *  const parsedVersion = parseLorawanMacVersion('1.1.0'); // returns 110
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
