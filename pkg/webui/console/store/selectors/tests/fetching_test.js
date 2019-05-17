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

import { createFetchingSelector } from '../fetching'

describe('fetching selectors', function () {
  const BASE_ACTION_TYPE = 'BASE_ACTION'
  let initialState = null

  describe('created with a single base action type', function () {
    const selector = createFetchingSelector(BASE_ACTION_TYPE)

    beforeAll(function () {
      initialState = { ui: { fetching: {}}}
    })

    describe('has no fetching entries', function () {
      it('should return `false`', function () {
        expect(selector(initialState)).toBe(false)
      })
    })

    describe('has fetching entry', function () {

      beforeAll(function () {
        initialState.ui.fetching[BASE_ACTION_TYPE] = true
      })

      it('should return `true`', function () {
        expect(selector(initialState)).toBe(true)
      })
    })
  })

  describe('created with two base action types', function () {
    const BASE_ACTION_TYPE_OTHER = 'BASE_ACTION_OTHER'
    const selector = createFetchingSelector([
      BASE_ACTION_TYPE,
      BASE_ACTION_TYPE_OTHER,
    ])

    beforeAll(function () {
      initialState = { ui: { fetching: {}}}
    })

    describe('has no fetching entries', function () {
      it('should return `false`', function () {
        expect(selector(initialState)).toBe(false)
      })
    })

    describe('has fetching entry', function () {

      beforeAll(function () {
        initialState.ui.fetching[BASE_ACTION_TYPE] = true
      })

      it('should return `true`', function () {
        expect(selector(initialState)).toBe(true)
      })

      describe('has two fetching entries', function () {
        beforeAll(function () {
          initialState.ui.fetching[BASE_ACTION_TYPE_OTHER] = true
        })

        it('should return `true`', function () {
          expect(selector(initialState)).toBe(true)
        })
      })
    })
  })
})
