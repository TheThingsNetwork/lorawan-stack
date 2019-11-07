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

import computePrefixes from './compute-prefix'

describe('computePrefix', function() {
  describe('should compoute prefix lengths that round exactly to characters', () => {
    const joinEUI = '1'.repeat(16)

    it('should compute empty prefix', () => {
      const length = 0
      const prefixes = computePrefixes(joinEUI, length)

      expect(prefixes).toHaveLength(1)
      expect(prefixes[0]).toBe('')
    })

    it('should compute prefixes with length rounded to full bytes (1-8)', () => {
      for (let i = 1; i <= 8; i++) {
        const length = i * 8
        const prefixes = computePrefixes(joinEUI, length)

        expect(prefixes).toHaveLength(1)
        expect(prefixes[0]).toBe(joinEUI.slice(0, length / 4))
      }
    })

    it('should compute prefixes with length rounded to full hex chars (e.g. 0.5, 1.5 bytes)', () => {
      for (let i = 1; i <= 8; i++) {
        const length = i * 8 - 4
        const prefixes = computePrefixes(joinEUI, length)

        expect(prefixes).toHaveLength(1)
        expect(prefixes[0]).toBe(joinEUI.slice(0, Math.max(length / 4, 1)))
      }
    })
  })

  describe('should compoute prefix lengths that do not round exactly to characters', () => {
    const data = [
      {
        joinEUI: '1111111111111111',
        length: 1,
        result: ['0', '1', '2', '3', '4', '5', '6', '7'],
      },
      {
        joinEUI: '1111111111111111',
        length: 2,
        result: ['0', '1', '2', '3'],
      },
      {
        joinEUI: '1111111111111111',
        length: 3,
        result: ['0', '1'],
      },
      {
        joinEUI: '1111111111111111',
        length: 5,
        result: ['10', '11', '12', '13', '14', '15', '16', '17'],
      },
      {
        joinEUI: '1111111111111111',
        length: 6,
        result: ['10', '11', '12', '13'],
      },
      {
        joinEUI: '1111111111111111',
        length: 7,
        result: ['10', '11'],
      },
      {
        joinEUI: '1111111111111111',
        length: 9,
        result: ['110', '111', '112', '113', '114', '115', '116', '117'],
      },
      {
        joinEUI: '1111111111111111',
        length: 10,
        result: ['110', '111', '112', '113'],
      },
    ]

    it('should compute prefix of lengths that do not round exactly to characters', () => {
      for (let i = 0; i < data.length; i++) {
        const { joinEUI, length, result } = data[i]
        const prefixes = computePrefixes(joinEUI, length)

        expect(prefixes).toStrictEqual(result)
      }
    })
  })
})
