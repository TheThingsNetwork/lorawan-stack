// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

/* eslint-env jest */
/* eslint-disable arrow-body-style */

import PropTypes from 'prop-types'
import check from './check-types'

test('check-types should throw when a type is wrong', () => {
  const types = {
    foo: PropTypes.string,
  }

  const val = {
    foo: 10,
  }

  expect(() => check(types, val, 'test', 'lib')).toThrow(/string/)
})

test('check-types should not throw when a type is not wrong', () => {
  const types = {
    foo: PropTypes.string,
  }

  const val = {
    foo: '10',
  }

  expect(() => check(types, val, 'test', 'lib')).not.toThrow()
})

test('check-types should throw when a type definition is wrong', () => {
  const types = {
    foo: 'invalid',
  }

  const val = {
    foo: 'ok',
  }

  expect(() => check(types, val, 'test', 'lib')).toThrow(/type `foo` is invalid/)
})

test('check-types should throw when a type definition throws', () => {
  const types = {
    foo () {
      throw new Error('huh?')
    },
  }

  const val = {
    foo: 'ok',
  }

  expect(() => check(types, val, 'test', 'lib')).toThrow(/Failed/)
})
