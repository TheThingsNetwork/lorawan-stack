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

import { trackerProxy, trackObject, removeDecorations } from './obj-tracking'

describe('trackerProxy', function() {
  let obj
  beforeEach(function() {
    obj = {
      a: {
        b: 'c',
        d: 'e',
      },
      e: 'f',
      _changed: [],
    }
  })

  test('tracks changes properly', function() {
    const proxyObj = new Proxy(obj, trackerProxy(obj))

    proxyObj.e = 'foo'

    expect(proxyObj._changed).toBeInstanceOf(Array)
    expect(proxyObj._changed).toHaveLength(1)
    expect(proxyObj._changed).toEqual(['e'])
  })

  test('tracks only marked properties', function() {
    const proxyObj = new Proxy(obj, trackerProxy({ a: 'b' }))

    proxyObj.e = 'foo'

    expect(proxyObj._changed).toHaveLength(0)
  })

  test('does not add tracked props more than once', function() {
    const proxyObj = new Proxy(obj, trackerProxy(obj))

    proxyObj.e = 'foo'
    proxyObj.e = 'foo'

    expect(proxyObj._changed).toEqual(['e'])
  })
})

describe('trackObject', function() {
  let obj
  beforeEach(function() {
    obj = {
      a: {
        b: {
          c: 'd',
        },
        e: 'f',
      },
      g: 'h',
    }
  })

  test('applies tracker proxies to all children', function() {
    obj = trackObject(obj)

    expect(obj.a.b._changed).toEqual([])
    expect(obj.a._changed).toEqual([])
    expect(obj._changed).toEqual([])

    obj.a.b.c = 'foo'
    obj.a.e = 'bar'
    obj.g = 'baz'

    expect(obj.a.b._changed).toEqual(['c'])
    expect(obj.a._changed).toEqual(['e'])
    expect(obj._changed).toEqual(['g'])
  })

  test('does not apply tracker inside arrays', function() {
    obj.i = ['j', 'k', 'l']
    obj = trackObject(obj)

    obj.i = ['foo']

    expect(obj.i).not.toContain('_changed')
    expect(obj._changed).toEqual(['i'])
  })
})

describe('removeDecorations', function() {
  let obj
  beforeEach(function() {
    obj = {
      a: {
        b: 'c',
        _changed: ['b'],
      },
      _changed: ['a'],
    }
  })

  test('removes all _changed properties', function() {
    removeDecorations(obj)

    expect(Object.keys(obj)).toEqual(['a'])
    expect(Object.keys(obj.a)).toEqual(['b'])
    expect(obj._changed).not.toBeDefined()
    expect(obj.a._changed).not.toBeDefined()
  })

  test('leaves source object untouched when clone parameter set', function() {
    const clonedObject = removeDecorations(obj, true)

    expect(clonedObject._changed).not.toBeDefined()
    expect(clonedObject.a.b).toBe('c')
    expect(obj._changed).toEqual(['a'])
    expect(obj.a._changed).toEqual(['b'])
  })
})
