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

/* eslint-env jest */
/* eslint-disable arrow-body-style */

import getByPath from './get-by-path'

test('flattens the object', () => {
  const testData = {
    a: {
      b: {
        c: 'foo',
      },
      d: 'bar',
    },
    e: 'baz',
    f: undefined,
    g: null,
    h: [1, 2, 3],
    i: {
      k: [3, 4, 5],
    },
  }

  expect(getByPath(testData, 'a.b.c')).toBe('foo')
  expect(getByPath(testData, 'a.d')).toBe('bar')
  expect(getByPath(testData, 'e')).toBe('baz')
  expect(getByPath(testData, 'f')).toBe(undefined)
  expect(getByPath(testData, 'g')).toBe(null)
  expect(getByPath(testData, 'h')).toMatchObject([1, 2, 3])
  expect(getByPath(testData, 'i.k')).toMatchObject([3, 4, 5])
  expect(getByPath(testData, 'i.e')).toBe(undefined)
  expect(getByPath(testData, 'a.b')).toMatchObject({ c: 'foo' })
})
