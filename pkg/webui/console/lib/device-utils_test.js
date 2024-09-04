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

import { parseLorawanMacVersion } from './device-utils'

describe('Parsing LoRaWAN Mac Version', () => {
  it.each([
    ['MAC_V1_0', 100],
    ['MAC_V1_0_1', 101],
    ['MAC_V1_0_2', 102],
    ['MAC_V1_0_3', 103],
    ['MAC_V1_0_4', 104],
    ['MAC_V1_1_0', 110],
    ['MAC_V1_2_0', 120],
    ['100', 0],
    ['101', 0],
    ['102', 0],
    ['103', 0],
    ['104', 0],
    ['110', 0],
    ['120', 0],
    [null, 0],
    [undefined, 0],
    ['invalid', 0],
    ['', 0],
  ])('yields parseLorawanVersion(%p) = %i', (actual, expected) => {
    expect(parseLorawanMacVersion(actual)).toBe(expected)
  })
})
