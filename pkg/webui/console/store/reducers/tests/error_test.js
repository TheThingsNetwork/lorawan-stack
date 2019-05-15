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

describe('error reducers', function () {
  const BASE = 'BASE_ACTION'
  const REQUEST = `${BASE}_REQUEST`
  const SUCCESS = `${BASE}_SUCCESS`
  const FAILURE = `${BASE}_FAILURE`

  it('should return the initial state', function () {
    expect(reducer(undefined, {})).toEqual({})
  })

  describe('dispatches invalid action type', function () {
    it('should ignore base type', function () {
      expect(reducer({}, { type: BASE })).toEqual({})
    })
  })

  describe('dispatches the `failure` action', function () {
    it('should set error', function () {
      const error = { status: 404 }
      expect(reducer({}, { type: FAILURE, error })).toEqual({
        [BASE]: error,
      })
    })

    describe('dispatches the `success` action', function () {
      it('should reset the error', function () {
        expect(reducer({}, { type: SUCCESS })).toEqual({
          [BASE]: undefined,
        })
      })
    })

    describe('dispatches the `request` action', function () {
      it('should reset the error', function () {
        expect(reducer({}, { type: REQUEST })).toEqual({
          [BASE]: undefined,
        })
      })
    })
  })
})
