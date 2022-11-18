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

import { hexToBase64, base64ToHex } from './bytes'

describe('Bytes utils', () => {
  const base64 = [
    'AQ==',
    'Ag==',
    '/w==',
    'vyUs3Q==',
    'MEijVA==',
    'EjRWeJCrze8=',
    'CZxkIN2eINE=',
    'AQIDBA==',
  ]
  const hex = [
    '01',
    '02',
    'ff',
    'bf252cdd',
    '3048a354',
    '1234567890abcdef',
    '099c6420dd9e20d1',
    '01 02 03 04 ',
  ]

  describe('when using base64ToHex', () => {
    const testTable = base64.map((value, index) => [value, hex[index]])
    it.each(testTable)('yields base64ToHex(%s) = %s', (base64Str, expectedHex) => {
      expect(base64ToHex(base64Str)).toBe(expectedHex.replace(/ /g, ''))
    })
  })

  describe('when using hexToBase64', () => {
    const testTable = hex.map((value, index) => [value, base64[index]])
    it.each(testTable)('yields hexToBase64(%s) = %s', (hexStr, expectedBase64) => {
      expect(hexToBase64(hexStr)).toBe(expectedBase64)
    })
  })
})
