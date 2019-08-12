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

import getCookieValue from './cookie'

describe('cookie utils', function() {
  describe('`document.cookie` is empty', function() {
    it('should return `undefined`', function() {
      const value = getCookieValue('missingKey')

      expect(value).toBeUndefined()
    })
  })

  describe('`document.cookie` has a single entry', function() {
    const key = 'testKey'
    const value = 'testValue'

    beforeEach(function() {
      document.cookie = `${key}=${value}`
    })

    afterEach(function() {
      document.cookie = `${key}=; expires=Thu, 01 Jan 1970 00:00:00 GMT`
    })

    it('should extract value for existing key', function() {
      expect(getCookieValue(key)).toBe(value)
    })

    it('should return `undefined` for non existing key', function() {
      expect(getCookieValue('nonExistingKey')).toBeUndefined()
    })
  })

  describe('`document.cookie` has multiple entries', function() {
    const key1 = 'testKey1'
    const key2 = 'testKey2'
    const value1 = 'testValue1'
    const value2 = 'testValue2'

    beforeEach(function() {
      document.cookie = `${key1}=${value1}`
      document.cookie = `${key2}=${value2}`
    })

    afterEach(function() {
      document.cookie = `${key1}=; expires=Thu, 01 Jan 1970 00:00:00 GMT`
      document.cookie = `${key2}=; expires=Thu, 01 Jan 1970 00:00:00 GMT`
    })

    it('should extract the value of the first entry', function() {
      expect(getCookieValue(key1)).toBe(value1)
    })

    it('should extract the value of the second entry', function() {
      expect(getCookieValue(key2)).toBe(value2)
    })
  })
})
