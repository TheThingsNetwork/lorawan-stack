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

import { createErrorSelector } from '@ttn-lw/lib/store/selectors/error'

describe('Error selectors', function () {
  const BASE_ACTION_TYPE = 'BASE_ACTION'
  let initialState = null

  describe('when created with a single base action type', function () {
    const selector = createErrorSelector(BASE_ACTION_TYPE)

    beforeAll(function () {
      initialState = { ui: { error: {} } }
    })

    describe('has no errors', function () {
      it('should return `undefined`', function () {
        expect(selector(initialState)).toBeUndefined()
      })
    })

    describe('has error', function () {
      const error = { status: 404 }

      beforeAll(function () {
        initialState.ui.error[BASE_ACTION_TYPE] = error
      })

      it('return the error object', function () {
        expect(selector(initialState)).toEqual(error)
      })
    })
  })

  describe('when created with two base action types', function () {
    const BASE_ACTION_TYPE_OTHER = 'BASE_ACTION_OTHER'
    const selector = createErrorSelector([BASE_ACTION_TYPE, BASE_ACTION_TYPE_OTHER])

    beforeAll(function () {
      initialState = { ui: { error: {} } }
    })

    describe('when there is no error', function () {
      it('return `undefined`', function () {
        expect(selector(initialState)).toBeUndefined()
      })
    })

    describe('when there is an error', function () {
      const not_found = { status: 404 }
      const forbidden = { status: 403 }

      beforeAll(function () {
        initialState.ui.error[BASE_ACTION_TYPE] = not_found
      })

      it('return the error object', function () {
        expect(selector(initialState)).toEqual(not_found)
      })

      describe('when there are two errors', function () {
        beforeAll(function () {
          initialState.ui.error[BASE_ACTION_TYPE_OTHER] = forbidden
        })

        it('return the first error object', function () {
          expect(selector(initialState)).toEqual(not_found)
        })
      })
    })
  })
})
