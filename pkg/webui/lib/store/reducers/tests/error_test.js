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

import reducer from '../ui/error'

describe('Error reducers', function () {
  const BASE = 'BASE_ACTION'
  const REQUEST = `${BASE}_REQUEST`
  const SUCCESS = `${BASE}_SUCCESS`
  const FAILURE = `${BASE}_FAILURE`
  const defaultState = {}

  it('return the initial state', function () {
    expect(reducer(undefined, {})).toEqual({})
  })

  describe('when dispatching an invalid action type', function () {
    it('ignores the base type', function () {
      expect(reducer(defaultState, { type: BASE })).toEqual(defaultState)
    })
  })

  describe('when dispatching the `failure` action', function () {
    const error = { status: 404 }
    let newState

    beforeAll(function () {
      newState = reducer(defaultState, { type: FAILURE, payload: error, error: true })
    })

    it('set the error', function () {
      expect(newState).toEqual({ [BASE]: error })
    })

    describe('when dispatching the `success` action', function () {
      it('reset the error', function () {
        expect(reducer(newState, { type: SUCCESS })).toEqual({
          [BASE]: undefined,
        })
      })
    })

    describe('when dispatching the `request` action', function () {
      it('reset the error', function () {
        expect(reducer(newState, { type: REQUEST })).toEqual({
          [BASE]: undefined,
        })
      })
    })
  })
})
