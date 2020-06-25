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

import toInputDate from './to-input-date'

describe('toInputDate', function() {
  describe('with correct argument', function() {
    const date = new Date('2020-09-24T12:00:00Z')

    it('should return a string length equal to 10', function() {
      expect(toInputDate(date)).toHaveLength(10)
    })

    it('should return a string of yyyy-mm--dd format', function() {
      expect(toInputDate(date)).toMatch(/([12]\d{3}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01]))/)
    })
  })

  describe('with incorrect argument', function() {
    const string = 'ABC123'
    const date = new Date('ABC123')

    it('random string should return undefined', function() {
      expect(toInputDate(string)).toBe(undefined)
    })

    it('invalid date should return undefined', function() {
      expect(toInputDate(date)).toBe(undefined)
    })
  })
})
