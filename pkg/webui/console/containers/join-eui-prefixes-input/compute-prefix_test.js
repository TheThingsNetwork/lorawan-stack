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

import computePrefix from './compute-prefix'

describe('computePrefix', function() {
  it('should compute empty prefix', function() {
    const joinEUI = '0'.repeat(16)
    const length = 0

    expect(computePrefix(joinEUI, length)).toBe('')
  })

  it('should compute prefix of 1 byte', function() {
    const joinEUI = '1'.repeat(16)
    const length = 8

    expect(computePrefix(joinEUI, length)).toBe('11')
  })

  it('should compute prefix of 1.5 bytes', function() {
    const joinEUI = '1'.repeat(16)
    const length = 12

    expect(computePrefix(joinEUI, length)).toBe('111')
  })

  it('should compute prefix of 8 bytes', function() {
    const joinEUI = '1'.repeat(16)
    const length = 64

    expect(computePrefix(joinEUI, length)).toBe(joinEUI)
  })

  it('should compute prefix of lengths that do not round exactly to characters', function() {
    const joinEUI = '123456789ABCDEF1'
    const lengths = Array.from(Array(65).keys()).filter(length => length % 4 !== 0) // consider only lengths that do not round exactly to a char
    const prefixes = Array.from(Array(17).keys())
      .slice(1, 17) // remove the empty string
      .map(i => joinEUI.slice(0, i)) // generate simplified prefixes

    for (let i = 0; i < lengths.length; i++) {
      const length = lengths[i]
      const prefix = prefixes[Math.floor(i / 3)]

      expect(computePrefix(joinEUI, length)).toBe(prefix)
    }
  })
})
