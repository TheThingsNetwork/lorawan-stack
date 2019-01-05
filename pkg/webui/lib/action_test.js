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

import PropTypes from 'prop-types'
import actions from './action'

test('should create the proper actions', () => {
  const ns = actions('ns', {
    foo: {
      types: {
        bar: PropTypes.string,
      },
    },
    bar: {
      baz: PropTypes.string,
      transform: PropTypes.number,
    },
    quu: {
      error: PropTypes.instanceOf(Error),
    },
    qux: {
      types: {
        bar: PropTypes.string,
      },
      transform (payload) {
        return {
          bar: (payload.bar || '').toString(),
        }
      },
    },
  })

  expect(ns).toHaveProperty('foo')
  expect(ns.foo).toHaveProperty('type', 'ns/foo')
  expect(ns.foo.toString()).toEqual('ns/foo')

  expect(ns).toHaveProperty('bar')
  expect(ns.bar).toHaveProperty('type', 'ns/bar')
  expect(ns.bar.toString()).toEqual('ns/bar')

  expect(() => ns.foo({ bar: 'bar' })).not.toThrow()
  expect(() => ns.foo({ bar: 10 })).toThrow()

  expect(() => ns.quu({ error: new Error() })).not.toThrow()
  expect(() => ns.qux({ bar: 10 })).not.toThrow()
})
